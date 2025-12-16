package worker

import (
	"context"
	"encoding/json"
	"fmt"
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

// 编码参数配置
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

func (h *VideoHandler) Handle(ctx context.Context, task *worker.Task) worker.Result {
	xlog.Msg("processing video task").Extra("taskId", task.Id).Infox(ctx)
	var req VideoProcessRequest
	if err := json.Unmarshal(task.InputArgs, &req); err != nil {
		return worker.NonRetryableResult(fmt.Errorf("invalid input: %w", err))
	}

	result, err := h.process(ctx, &req)
	if err != nil {
		xlog.Msg("video process failed").Err(err).Extra("taskId", task.Id).Errorx(ctx)
		return worker.NonRetryableResult(err)
	}

	output, _ := json.Marshal(result)
	return worker.SuccessResult(json.RawMessage(output))
}

func (h *VideoHandler) process(ctx context.Context, req *VideoProcessRequest) (*VideoProcessResponse, error) {
	inputURL, err := h.resolveInputURL(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("resolve input url failed: %w", err)
	}
	if inputURL == "" {
		return nil, fmt.Errorf("input url is empty")
	}

	resp := &VideoProcessResponse{}

	for _, out := range req.Outputs {
		opts := h.buildOptions(out.Settings)
		// 这里会开进程处理
		result, err := h.processor.ProcessSingle(ctx, ffmpeg.SingleProcessRequest{
			InputURL:  inputURL,
			Bucket:    out.Bucket,
			OutputKey: out.OutputKey,
			Options:   opts,
		})
		if err != nil {
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
	}

	if req.Thumbnail != nil && req.Thumbnail.OutputKey != "" {
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
	}

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

func (h *VideoHandler) buildOptions(s *EncodeSettings) []ffmpeg.OptionFunc {
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
	if s.MaxHeight > 0 {
		opts = append(opts, ffmpeg.WithMaxHeight(s.MaxHeight))
	}
	if s.MaxWidth > 0 {
		opts = append(opts, ffmpeg.WithMaxWidth(s.MaxWidth))
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
