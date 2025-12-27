package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/storage"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type Processor struct {
	ff      *FFmpeg
	storage *storage.Storage
}

func NewProcessor(ff *FFmpeg, storage *storage.Storage) *Processor {
	return &Processor{
		ff:      ff,
		storage: storage,
	}
}

type SingleProcessRequest struct {
	InputURL        string
	Bucket          string
	OutputKey       string
	Options         []OptionFunc
	UseStreamUpload bool // 是否使用流式上传，false（默认）使用临时文件+faststart，true 使用流式+fragmented MP4
}

// ProcessResult 转码结果，包含输出视频的实际参数
type ProcessResult struct {
	// 视频流信息
	Width     int     // 视频宽度
	Height    int     // 视频高度
	Duration  float64 // 视频时长（秒）
	Bitrate   int64   // 总码率（bps）
	Codec     string  // 视频编码器
	Framerate float64 // 帧率

	// 音频流信息
	AudioCodec      string // 音频编码器
	AudioSampleRate int    // 采样率（Hz）
	AudioChannels   int    // 声道数
	AudioBitrate    int64  // 音频码率（bps）
}

func (p *Processor) ProcessSingle(ctx context.Context, req SingleProcessRequest) (*ProcessResult, error) {
	opt := applyOptions(req.Options...)

	// 根据配置选择上传方式
	if req.UseStreamUpload {
		// 方式一：流式上传（fragmented MP4，适合流媒体场景）
		return p.processWithStreamUpload(ctx, req.InputURL, req.Bucket, req.OutputKey, opt)
	}

	// 方式二（默认）：临时文件 + faststart（适合浏览器渐进式下载）
	return p.processWithFileUpload(ctx, req.InputURL, req.Bucket, req.OutputKey, opt)
}

// processWithStreamUpload 流式上传方式：边转码边上传
// 使用 fragmented MP4 格式，适合 HLS/DASH 流媒体场景
func (p *Processor) processWithStreamUpload(ctx context.Context, inputURL, bucket, outputKey string, opt *Option) (*ProcessResult, error) {
	reader, err := p.ff.Transcode(ctx, TranscodeInput{
		InputURL: inputURL,
		Option:   opt,
	})
	if err != nil {
		return nil, fmt.Errorf("transcode failed: %w", err)
	}

	uploadErr := p.storage.UploadStream(ctx, bucket, outputKey, reader, "video/mp4")

	closeErr := reader.Close() // block here, wait for ffmpeg to finish
	if closeErr != nil {
		return nil, fmt.Errorf("ffmpeg failed: %w", closeErr)
	}
	if uploadErr != nil {
		return nil, fmt.Errorf("upload failed: %w", uploadErr)
	}

	return p.getVideoMetadata(ctx, bucket, outputKey)
}

// processWithFileUpload 临时文件上传方式：先转码到文件，再上传
// 使用 faststart 优化，moov atom 在文件头，支持浏览器渐进式下载
func (p *Processor) processWithFileUpload(ctx context.Context, inputURL, bucket, outputKey string, opt *Option) (*ProcessResult, error) {
	// 创建临时文件
	tmpFile := filepath.Join(os.TempDir(), uuid.New().String()+".mp4")
	
	// 确保临时文件被删除
	var uploadSuccess bool
	defer func() {
		if err := os.Remove(tmpFile); err != nil && !os.IsNotExist(err) {
			xlog.Msg("failed to remove temp file").
				Extra("file", tmpFile).
				Extra("upload_success", uploadSuccess).
				Err(err).Errorx(ctx)
		}
	}()

	// 转码到临时文件（使用 faststart）
	if err := p.ff.TranscodeToFile(ctx, inputURL, tmpFile, opt); err != nil {
		return nil, fmt.Errorf("transcode failed: %w", err)
	}

	// 上传文件到存储
	if err := p.storage.UploadFile(ctx, bucket, outputKey, tmpFile, "video/mp4"); err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	uploadSuccess = true

	return p.getVideoMetadata(ctx, bucket, outputKey)
}

// getVideoMetadata 获取已上传视频的元数据
func (p *Processor) getVideoMetadata(ctx context.Context, bucket, outputKey string) (*ProcessResult, error) {
	url, err := p.storage.GetPresignedURL(ctx, bucket, outputKey, time.Hour)
	if err != nil {
		xlog.Msg("get presigned url failed").Err(err).Errorx(ctx)
		return &ProcessResult{}, nil
	}

	probeResult, err := Probe(ctx, url)
	if err != nil {
		xlog.Msg("probe failed").Err(err).Errorx(ctx)
		return &ProcessResult{}, nil
	}

	return &ProcessResult{
		Width:           probeResult.Width,
		Height:          probeResult.Height,
		Duration:        probeResult.Duration,
		Bitrate:         probeResult.Bitrate,
		Codec:           probeResult.Codec,
		Framerate:       probeResult.Framerate,
		AudioCodec:      probeResult.AudioCodec,
		AudioSampleRate: probeResult.AudioSampleRate,
		AudioChannels:   probeResult.AudioChannels,
		AudioBitrate:    probeResult.AudioBitrate,
	}, nil
}

func (p *Processor) ExtractAndUploadThumbnail(ctx context.Context, bucket, inputURL, outputKey string, atSecond float64) error {
	filePath, cleanup, err := p.ff.ExtractThumbnail(ctx, inputURL, atSecond)
	if err != nil {
		return err
	}
	defer cleanup()

	return p.storage.UploadFile(ctx, bucket, outputKey, filePath, "image/jpeg")
}
