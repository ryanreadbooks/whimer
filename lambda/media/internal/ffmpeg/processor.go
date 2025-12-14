package ffmpeg

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/lambda/media/internal/storage"
)

type Processor struct {
	ff      *FFmpeg
	storage *storage.Storage
	stream  bool
}

func NewProcessor(ff *FFmpeg, storage *storage.Storage, stream bool) *Processor {
	return &Processor{
		ff:      ff,
		storage: storage,
		stream:  stream,
	}
}

type SingleProcessRequest struct {
	InputURL  string
	Bucket    string
	OutputKey string
	Options   []OptionFunc
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

	mode := OutputModeFile
	if p.stream {
		mode = OutputModeStream
	}

	output, err := p.ff.Transcode(ctx, TranscodeInput{
		InputURL: req.InputURL,
		Option:   opt,
	}, mode)
	if err != nil {
		return nil, fmt.Errorf("transcode failed: %w", err)
	}

	// 落盘模式下，先 probe 获取实际参数，再上传
	var probeResult *ProbeResult
	if output.FilePath != "" {
		probeResult, err = Probe(ctx, output.FilePath)
		if err != nil {
			if output.Cleanup != nil {
				output.Cleanup()
			}
			return nil, fmt.Errorf("probe output failed: %w", err)
		}
	}

	if output.Cleanup != nil {
		defer output.Cleanup()
	}

	err = p.storage.UploadFromOutput(ctx, req.Bucket, req.OutputKey, output.FilePath, output.Reader, "video/mp4")
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	// 流式模式下，上传后再 probe（从存储获取）
	if probeResult == nil {
		url := p.storage.GetObjectURL(req.Bucket, req.OutputKey)
		probeResult, err = Probe(ctx, url)
		if err != nil {
			// probe 失败不影响整体流程，返回空结果
			return &ProcessResult{}, nil
		}
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
