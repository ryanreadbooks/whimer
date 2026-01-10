package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/worker"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/ffmpeg"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/storage"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

// InputMode 输入模式
type InputMode string

const (
	// 通过 URL 获取输入视频
	InputModeURL InputMode = "url"
	// s3 / minio
	InputModeS3 InputMode = "s3"
)

// VideoProcessRequest 视频处理请求
type VideoProcessRequest struct {
	// 输入模式: url / s3，默认 url
	InputMode InputMode `json:"input_mode,omitempty"`

	// 输入视频地址 (input_mode=url 时使用)
	InputURL string `json:"input_url,omitempty"`

	// 输入存储桶 (input_mode=s3 时使用)
	InputBucket string `json:"input_bucket,omitempty"`

	// 输入文件路径 (input_mode=s3 时使用)
	InputKey string `json:"input_key,omitempty"`

	// 输出配置列表，支持同时输出多个不同规格的视频
	Outputs []OutputConfig `json:"outputs"`

	// 缩略图配置，可选
	Thumbnail *ThumbnailConfig `json:"thumbnail,omitempty"`
}

type OutputConfig struct {
	// 输出存储桶名称
	Bucket string `json:"bucket"`

	// 输出文件路径/键名
	OutputKey string `json:"output_key"`

	// 编码参数，不传则使用默认值 (H.264, CRF 23)
	Settings *EncodeSettings `json:"settings,omitempty"`
}

// EncodeSettings 编码参数配置
type EncodeSettings struct {
	// VideoCodec 视频编码器:
	//   - libx264: H.264
	//   - libx265: H.265/HEVC
	//   - libsvtav1: AV1
	//   - copy: 不转码
	VideoCodec string `json:"video_codec,omitempty"`

	// 音频编码器: aac, copy(不转码)
	AudioCodec string `json:"audio_codec,omitempty"`

	// 视频目标码率，如 "1500k", "2M"，CRF 模式下仅作参考
	VideoBitrate string `json:"video_bitrate,omitempty"`

	// 音频码率，如 "128k", "192k"
	AudioBitrate string `json:"audio_bitrate,omitempty"`

	// 最大高度（像素），保持宽高比缩放，不会放大
	MaxHeight int `json:"max_height,omitempty"`

	// 最大宽度（像素），保持宽高比缩放，不会放大
	MaxWidth int `json:"max_width,omitempty"`

	// Preset 编码速度预设:
	//   - H.264/H.265: ultrafast/fast/medium/slow/veryslow，越慢压缩率越高
	//   - AV1: 0-13 (数字字符串)，越小越慢质量越好，推荐 "6"-"10"
	Preset string `json:"preset,omitempty"`

	// 质量因子，值越小质量越高:
	//   - H.264/H.265: 0-51，推荐 18-28，默认 23
	//   - AV1: 0-63，推荐 28-42，默认 35
	CRF int `json:"crf,omitempty"`

	// 目标帧率，0 表示保持原帧率
	Framerate int `json:"framerate,omitempty"`

	// Auto720p 智能分辨率适配模式
	// 开启后会自动检测源视频分辨率并智能决定是否缩放:
	//   - 标准分辨率 (1080p/720p/480p等): 不处理，保持原样
	//   - 超过 1080p 的非标准分辨率: 缩放到 1080p
	//   - 720p ~ 1080p 之间的非标准分辨率: 缩放到 720p
	//   - 小于 720p: 不处理
	// 此选项会覆盖 MaxHeight/MaxWidth 设置
	Auto720p bool `json:"auto_720p,omitempty"`
}

// 缩略图配置
type ThumbnailConfig struct {
	// 输出存储桶名称
	Bucket string `json:"bucket"`

	// 输出文件路径/键名
	OutputKey string `json:"output_key"`

	// 截取时间点（秒），默认 1.0 秒
	AtSecond float64 `json:"at_second,omitempty"`
}

// 视频处理响应
type VideoProcessResponse struct {
	// 输出结果列表
	Outputs []VideoOutput `json:"outputs"`

	// 缩略图结果
	Thumbnail *ThumbnailOutput `json:"thumbnail,omitempty"`
}

// 单个输出结果
type VideoOutput struct {
	// 桶名称
	Bucket string `json:"bucket"`

	// key
	Key string `json:"key"`

	// 输出视频的实际参数（通过 ffprobe 获取）
	Info *VideoInfo `json:"info"`
}

// 输出视频的实际参数
type VideoInfo struct {
	// 视频宽度（像素）
	Width int `json:"width"`

	// 视频高度（像素）
	Height int `json:"height"`

	// 视频时长（秒）
	Duration float64 `json:"duration"`

	// 总码率（bps）
	Bitrate int64 `json:"bitrate"`

	// 视频编码器
	Codec string `json:"codec"`

	// 帧率
	Framerate float64 `json:"framerate"`

	// 音频编码器
	AudioCodec string `json:"audio_codec"`

	// 音频采样率（Hz）
	AudioSampleRate int `json:"audio_sample_rate"`

	// 音频声道数
	AudioChannels int `json:"audio_channels"`

	// 音频码率（bps）
	AudioBitrate int64 `json:"audio_bitrate"`
}

// ThumbnailOutput 缩略图输出结果
type ThumbnailOutput struct {
	// 存储桶名称
	Bucket string `json:"bucket"`

	// 文件路径/键名
	Key string `json:"key"`

	// 截取时间点（秒）
	AtSecond float64 `json:"at_second"`
}

type VideoHandler struct {
	processor *ffmpeg.Processor
	storage   *storage.Storage
}

func NewVideoHandler(processor *ffmpeg.Processor, storage *storage.Storage) *VideoHandler {
	return &VideoHandler{processor: processor, storage: storage}
}

// Handle 处理视频任务，支持进度上报和中断检测
func (h *VideoHandler) Handle(tc worker.TaskContext) worker.Result {
	ctx := tc.Context()
	task := tc.Task()

	xlog.Msg("processing video task").Extra("taskId", task.Id).Infox(ctx)

	var req VideoProcessRequest
	if err := json.Unmarshal(task.InputArgs, &req); err != nil {
		return worker.NonRetryableResult(fmt.Errorf("invalid input: %w", err))
	}

	// 设置进度提供者，用于心跳自动上报进度
	var progress atomic.Int64
	tc.SetProgressProvider(worker.ProgressFunc(func() int64 {
		return progress.Load()
	}))

	result, err := h.process(ctx, &req, tc, &progress)
	if err != nil {
		// 如果是任务被中断，返回特定错误
		if tc.IsAborted() {
			xlog.Msg("video task aborted").Extra("taskId", task.Id).Infox(ctx)
			return worker.Result{Error: worker.ErrTaskAborted}
		}
		xlog.Msg("video process failed").Err(err).Extra("taskId", task.Id).Errorx(ctx)
		return worker.NonRetryableResult(err)
	}

	output, _ := json.Marshal(result)
	return worker.SuccessResult(json.RawMessage(output))
}

func (h *VideoHandler) process(
	ctx context.Context,
	req *VideoProcessRequest,
	tc worker.TaskContext,
	progress *atomic.Int64,
) (*VideoProcessResponse, error) {
	// 检查是否已中断
	if tc.IsAborted() {
		return nil, worker.ErrTaskAborted
	}

	inputURL, err := h.resolveInputURL(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("resolve input url failed: %w", err)
	}
	if inputURL == "" {
		return nil, fmt.Errorf("input url is empty")
	}

	// 计算总步骤数用于进度计算
	totalSteps := len(req.Outputs)
	if req.Thumbnail != nil && req.Thumbnail.OutputKey != "" {
		totalSteps++
	}
	currentStep := 0

	// 预先 probe 获取源视频信息（用于 Auto720p 判断）
	var srcProbe *ffmpeg.ProbeResult
	if h.needsProbeForAuto720p(req.Outputs) {
		srcProbe, err = ffmpeg.Probe(ctx, inputURL)
		if err != nil {
			xlog.Msg("probe source video failed, will skip auto720p optimization").Err(err).Infox(ctx)
		}
	}

	resp := &VideoProcessResponse{}

	for _, out := range req.Outputs {
		// 每个输出前检查是否中断
		if tc.IsAborted() {
			return nil, worker.ErrTaskAborted
		}

		opts := h.buildOptionsWithProbe(out.Settings, srcProbe)
		// 这里会开进程处理
		result, err := h.processor.ProcessSingle(ctx, ffmpeg.SingleProcessRequest{
			InputURL:  inputURL,
			Bucket:    out.Bucket,
			OutputKey: out.OutputKey,
			Options:   opts,
		})
		if err != nil {
			// context 取消时返回中断错误
			if ctx.Err() != nil {
				return nil, worker.ErrTaskAborted
			}
			return nil, fmt.Errorf("process %s failed: %w", out.OutputKey, err)
		}

		resp.Outputs = append(resp.Outputs, VideoOutput{
			Bucket: out.Bucket,
			Key:    out.OutputKey,
			Info: &VideoInfo{
				Width:           result.Width,
				Height:          result.Height,
				Duration:        result.Duration,
				Bitrate:         result.Bitrate,
				Codec:           result.Codec,
				Framerate:       result.Framerate,
				AudioCodec:      result.AudioCodec,
				AudioSampleRate: result.AudioSampleRate,
				AudioChannels:   result.AudioChannels,
				AudioBitrate:    result.AudioBitrate,
			},
		})

		// 更新进度
		currentStep++
		progress.Store(int64(currentStep * 100 / totalSteps))
	}

	if req.Thumbnail != nil && req.Thumbnail.OutputKey != "" {
		// 缩略图前检查是否中断
		if tc.IsAborted() {
			return nil, worker.ErrTaskAborted
		}

		atSecond := req.Thumbnail.AtSecond
		if atSecond <= 0 {
			atSecond = 1.0
		}
		if err := h.processor.ExtractAndUploadThumbnail(
			ctx,
			req.Thumbnail.Bucket,
			inputURL,
			req.Thumbnail.OutputKey,
			atSecond,
		); err != nil {
			xlog.Msg("extract thumbnail failed").Err(err).Infox(ctx)
		} else {
			resp.Thumbnail = &ThumbnailOutput{
				Bucket:   req.Thumbnail.Bucket,
				Key:      req.Thumbnail.OutputKey,
				AtSecond: atSecond,
			}
		}

		// 更新进度
		currentStep++
		progress.Store(int64(currentStep * 100 / totalSteps))
	}

	// 完成
	progress.Store(100)
	return resp, nil
}

func (h *VideoHandler) resolveInputURL(ctx context.Context, req *VideoProcessRequest) (string, error) {
	switch req.InputMode {
	case InputModeS3:
		// 生成 24小时有效的签名URL
		return h.storage.GetPresignedURL(ctx, req.InputBucket, req.InputKey, 24*time.Hour)
	default:
		return req.InputURL, nil
	}
}

// needsProbeForAuto720p 检查是否有任何输出配置启用了 Auto720p
func (h *VideoHandler) needsProbeForAuto720p(outputs []OutputConfig) bool {
	for _, out := range outputs {
		if out.Settings != nil && out.Settings.Auto720p {
			return true
		}
	}
	return false
}

// buildOptionsWithProbe 根据 probe 结果和设置构建 ffmpeg 选项
func (h *VideoHandler) buildOptionsWithProbe(s *EncodeSettings, probe *ffmpeg.ProbeResult) []ffmpeg.OptionFunc {
	if s == nil {
		return nil
	}

	var opts []ffmpeg.OptionFunc

	if s.VideoCodec != "" {
		opts = append(opts, ffmpeg.WithVideoCodec(ffmpeg.VideoCodec(s.VideoCodec)))
	}
	if s.AudioCodec != "" {
		opts = append(opts, ffmpeg.WithAudioCodec(ffmpeg.AudioCodec(s.AudioCodec)))
	}
	if s.VideoBitrate != "" {
		opts = append(opts, ffmpeg.WithVideoBitrate(s.VideoBitrate))
	}
	if s.AudioBitrate != "" {
		opts = append(opts, ffmpeg.WithAudioBitrate(s.AudioBitrate))
	}

	// Auto720p 智能缩放处理
	if s.Auto720p && probe != nil {
		// 根据源分辨率智能决定是否缩放
		// - 标准分辨率（1080p/720p等）：不处理
		// - 非标准分辨率：缩放到推荐的标准分辨率
		if targetDim, needsScale := ffmpeg.RecommendedScaleTarget(probe.Width, probe.Height); needsScale {
			opts = append(opts, ffmpeg.WithMaxHeight(targetDim))
		}
	} else {
		// 非 Auto720p 模式，使用原有逻辑
		if s.MaxHeight > 0 {
			opts = append(opts, ffmpeg.WithMaxHeight(s.MaxHeight))
		}
		if s.MaxWidth > 0 {
			opts = append(opts, ffmpeg.WithMaxWidth(s.MaxWidth))
		}
	}

	if s.Preset != "" {
		opts = append(opts, ffmpeg.WithPreset(s.Preset))
	}
	if s.CRF > 0 {
		opts = append(opts, ffmpeg.WithCRF(s.CRF))
	}
	if s.Framerate > 0 {
		opts = append(opts, ffmpeg.WithFramerate(s.Framerate))
	}

	return opts
}

func (h *VideoHandler) buildOptions(s *EncodeSettings) []ffmpeg.OptionFunc {
	return h.buildOptionsWithProbe(s, nil)
}
