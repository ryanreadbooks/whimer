package ffmpeg

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

// ProbeResult 视频探测结果
type ProbeResult struct {
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

type probeFormat struct {
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
}

type probeStream struct {
	CodecType    string `json:"codec_type"`
	CodecName    string `json:"codec_name"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	RFrameRate   string `json:"r_frame_rate"`
	AvgFrameRate string `json:"avg_frame_rate"`
	SampleRate   string `json:"sample_rate"`
	Channels     int    `json:"channels"`
	BitRate      string `json:"bit_rate"`
}

type probeOutput struct {
	Format  probeFormat   `json:"format"`
	Streams []probeStream `json:"streams"`
}

func Probe(ctx context.Context, input string) (*ProbeResult, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		input,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var probe probeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, err
	}

	result := &ProbeResult{}
	result.Duration, _ = strconv.ParseFloat(probe.Format.Duration, 64)
	result.Bitrate, _ = strconv.ParseInt(probe.Format.BitRate, 10, 64)

	for _, stream := range probe.Streams {
		switch stream.CodecType {
		case "video":
			result.Width = stream.Width
			result.Height = stream.Height
			result.Codec = stream.CodecName
			result.Framerate = parseFramerate(stream.RFrameRate)
		case "audio":
			result.AudioCodec = stream.CodecName
			result.AudioSampleRate, _ = strconv.Atoi(stream.SampleRate)
			result.AudioChannels = stream.Channels
			result.AudioBitrate, _ = strconv.ParseInt(stream.BitRate, 10, 64)
		}
	}

	return result, nil
}

func parseFramerate(s string) float64 {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0
	}
	num, _ := strconv.ParseFloat(parts[0], 64)
	den, _ := strconv.ParseFloat(parts[1], 64)
	if den == 0 {
		return 0
	}
	return num / den
}
