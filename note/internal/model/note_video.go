package model

import (
	"fmt"
	"path/filepath"
	"strings"
)

type NoteVideoMedia struct {
	Width        int32  `json:"width"`
	Height       int32  `json:"height"`
	VideoCodec   string `json:"video_codec"`
	VideoBitrate int32  `json:"video_bitrate"`
	FrameRate    int32  `json:"frame_rate"`
	Duration     int32  `json:"duration"`
	Format       string `json:"format"`
	AudioCodec   string `json:"audio_codec"`
	AudioBitrate int32  `json:"audio_bitrate"`
}

type NoteVideoItem struct {
	Key   string          `json:"key"` // 资源key 包含bucket名
	Media *NoteVideoMedia `json:"meta"`

	bucket string `json:"-"` // 非全场景必须字段 用到时手动Set
}

func (v *NoteVideoItem) SetBucket(bucket string) {
	if v == nil {
		return
	}
	v.bucket = bucket
}

func (v *NoteVideoItem) GetBucket() string {
	if v == nil {
		return ""
	}
	return v.bucket
}

func (v *NoteVideoItem) TrimBucket() string {
	if v == nil {
		return ""
	}

	if strings.HasPrefix(v.Key, v.GetBucket()+"/") {
		return strings.TrimPrefix(v.Key, v.GetBucket()+"/")
	}

	return v.Key
}

type SupportedVideoSuffix string

const (
	SupportedVideoH264Suffix SupportedVideoSuffix = "_264"
	SupportedVideoH265Suffix SupportedVideoSuffix = "_265"
	SupportedVideoAV1Suffix  SupportedVideoSuffix = "_av1"
)

// 笔记视频资源
type NoteVideo struct {
	H264 *NoteVideoItem `json:"h264,omitempty"`
	H265 *NoteVideoItem `json:"h265,omitempty"`
	AV1  *NoteVideoItem `json:"av1,omitempty"`

	rawUrl    string `json:"-"` // 非必要字段 需要时填充
	rawBucket string `json:"-"` // 非必要字段 需要时填充
}

func (v *NoteVideo) GetRawUrl() string {
	if v == nil {
		return ""
	}
	return v.rawUrl
}

func (v *NoteVideo) SetRawUrl(url string) {
	if v == nil {
		return
	}
	v.rawUrl = url
}

func (v *NoteVideo) SetRawBucket(bucket string) {
	if v == nil {
		return
	}
	v.rawBucket = bucket
}

func (v *NoteVideo) GetRawBucket() string {
	if v == nil {
		return ""
	}
	return v.rawBucket
}

func (v *NoteVideo) SetTargetBucket(bucket string) {
	if v == nil {
		return
	}
	v.H264.SetBucket(bucket)
	v.H265.SetBucket(bucket)
	v.AV1.SetBucket(bucket)
}

func FormatNoteVideoKey(id string, suffix SupportedVideoSuffix) string {
	ext := filepath.Ext(id)
	basename := id
	if ext != "" {
		basename = strings.TrimSuffix(id, ext)
	}

	return fmt.Sprintf("%s%s.mp4", basename, suffix)
}

// 视频的metadata
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

type VideoAssetMetadata struct {
	Key string
	Info *VideoInfo
}