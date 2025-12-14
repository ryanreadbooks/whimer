package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/worker"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/ffmpeg"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

// VideoProcessRequest 视频处理请求
type VideoProcessRequest struct {
	// InputURL 输入视频地址，支持 http/https/本地路径
	InputURL string `json:"input_url"`

	// Outputs 输出配置列表，支持同时输出多个不同规格的视频
	Outputs []OutputConfig `json:"outputs"`

	// Thumbnail 缩略图配置，可选
	Thumbnail *ThumbnailConfig `json:"thumbnail,omitempty"`
}

// OutputConfig 单个输出配置
type OutputConfig struct {
	// Bucket 输出存储桶名称
	Bucket string `json:"bucket"`

	// OutputKey 输出文件路径/键名
	OutputKey string `json:"output_key"`

	// Settings 编码参数，不传则使用默认值 (H.264, CRF 23)
	Settings *EncodeSettings `json:"settings,omitempty"`
}

// EncodeSettings 编码参数配置
type EncodeSettings struct {
	// VideoCodec 视频编码器: libx264(H.264), libx265(H.265), copy(不转码)
	VideoCodec string `json:"video_codec,omitempty"`

	// AudioCodec 音频编码器: aac, copy(不转码)
	AudioCodec string `json:"audio_codec,omitempty"`

	// VideoBitrate 视频目标码率，如 "1500k", "2M"，CRF 模式下仅作参考
	VideoBitrate string `json:"video_bitrate,omitempty"`

	// AudioBitrate 音频码率，如 "128k", "192k"
	AudioBitrate string `json:"audio_bitrate,omitempty"`

	// MaxHeight 最大高度（像素），保持宽高比缩放，不会放大
	MaxHeight int `json:"max_height,omitempty"`

	// MaxWidth 最大宽度（像素），保持宽高比缩放，不会放大
	MaxWidth int `json:"max_width,omitempty"`

	// Preset 编码速度预设: ultrafast/fast/medium/slow/veryslow，越慢压缩率越高
	Preset string `json:"preset,omitempty"`

	// CRF 质量因子 (0-51)，值越小质量越高，推荐 18-28，默认 23
	CRF int `json:"crf,omitempty"`

	// Framerate 目标帧率，0 表示保持原帧率
	Framerate int `json:"framerate,omitempty"`
}

// ThumbnailConfig 缩略图配置
type ThumbnailConfig struct {
	// Bucket 输出存储桶名称
	Bucket string `json:"bucket"`

	// OutputKey 输出文件路径/键名
	OutputKey string `json:"output_key"`

	// AtSecond 截取时间点（秒），默认 1.0 秒
	AtSecond float64 `json:"at_second,omitempty"`
}

// VideoProcessResponse 视频处理响应
type VideoProcessResponse struct {
	// Outputs 输出结果列表
	Outputs []VideoOutput `json:"outputs"`

	// Thumbnail 缩略图结果
	Thumbnail *ThumbnailOutput `json:"thumbnail,omitempty"`
}

// VideoOutput 单个输出结果
type VideoOutput struct {
	// Bucket 存储桶名称
	Bucket string `json:"bucket"`

	// Key 文件路径/键名
	Key string `json:"key"`

	// Info 输出视频的实际参数（通过 ffprobe 获取）
	Info *VideoInfo `json:"info"`
}

// VideoInfo 输出视频的实际参数
type VideoInfo struct {
	// Width 视频宽度（像素）
	Width int `json:"width"`

	// Height 视频高度（像素）
	Height int `json:"height"`

	// Duration 视频时长（秒）
	Duration float64 `json:"duration"`

	// Bitrate 总码率（bps）
	Bitrate int64 `json:"bitrate"`

	// Codec 视频编码器
	Codec string `json:"codec"`

	// Framerate 帧率
	Framerate float64 `json:"framerate"`

	// AudioCodec 音频编码器
	AudioCodec string `json:"audio_codec"`

	// AudioSampleRate 音频采样率（Hz）
	AudioSampleRate int `json:"audio_sample_rate"`

	// AudioChannels 音频声道数
	AudioChannels int `json:"audio_channels"`

	// AudioBitrate 音频码率（bps）
	AudioBitrate int64 `json:"audio_bitrate"`
}

// ThumbnailOutput 缩略图输出结果
type ThumbnailOutput struct {
	// Bucket 存储桶名称
	Bucket string `json:"bucket"`

	// Key 文件路径/键名
	Key string `json:"key"`

	// AtSecond 截取时间点（秒）
	AtSecond float64 `json:"at_second"`
}

type VideoHandler struct {
	processor *ffmpeg.Processor
}

func NewVideoHandler(processor *ffmpeg.Processor) *VideoHandler {
	return &VideoHandler{processor: processor}
}

func (h *VideoHandler) Handle(ctx context.Context, task *worker.Task) worker.Result {
	xlog.Msg("processing video task").Extra("taskId", task.Id).Infox(ctx)

	var req VideoProcessRequest
	if err := json.Unmarshal(task.InputArgs, &req); err != nil {
		return worker.Result{Error: fmt.Errorf("invalid input: %w", err)}
	}

	result, err := h.process(ctx, &req)
	if err != nil {
		xlog.Msg("video process failed").Err(err).Extra("taskId", task.Id).Errorx(ctx)
		return worker.Result{Error: err}
	}

	output, _ := json.Marshal(result)
	return worker.Result{Output: string(output)}
}

func (h *VideoHandler) process(ctx context.Context, req *VideoProcessRequest) (*VideoProcessResponse, error) {
	resp := &VideoProcessResponse{}

	for _, out := range req.Outputs {
		opts := h.buildOptions(out.Settings)

		result, err := h.processor.ProcessSingle(ctx, ffmpeg.SingleProcessRequest{
			InputURL:  req.InputURL,
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
			req.InputURL,
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
