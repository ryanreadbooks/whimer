package ffmpeg

type VideoCodec string
type AudioCodec string
type OutputFormat string

const (
	VideoCodecH264 VideoCodec = "libx264" // H.264 编码，兼容性最好
	VideoCodecH265 VideoCodec = "libx265" // H.265/HEVC 编码，压缩率更高但兼容性较差
	VideoCodecCopy VideoCodec = "copy"    // 不重新编码，直接复制

	AudioCodecAAC  AudioCodec = "aac"  // AAC 音频编码
	AudioCodecCopy AudioCodec = "copy" // 不重新编码，直接复制

	OutputFormatMP4 OutputFormat = "mp4" // MP4 容器格式
	OutputFormatHLS OutputFormat = "hls" // HLS 流媒体格式
)

// Option 转码参数配置
type Option struct {
	VideoCodec   VideoCodec   // 视频编码器 (libx264/libx265/copy)
	AudioCodec   AudioCodec   // 音频编码器 (aac/copy)
	OutputFormat OutputFormat // 输出格式 (mp4/hls)
	VideoBitrate string       // 视频目标码率 (如 "1500k", "2M")，与 CRF 同时存在时 CRF 优先
	AudioBitrate string       // 音频码率 (如 "128k", "192k")
	MaxHeight    int          // 最长边限制，横屏限宽/竖屏限高，保持宽高比，只缩不放
	MaxWidth     int          // 同 MaxHeight，优先使用较小值
	Preset       string       // 编码速度预设 (ultrafast/fast/medium/slow/veryslow)，越慢压缩率越高
	CRF          int          // 质量因子 (0-51)，值越小质量越高，推荐 18-28，默认 23
	Framerate    int          // 目标帧率，0 表示保持原帧率
	ExtraArgs    []string     // 额外的 ffmpeg 参数
}

type OptionFunc func(*Option)

func WithVideoCodec(codec VideoCodec) OptionFunc {
	return func(o *Option) { o.VideoCodec = codec }
}

func WithAudioCodec(codec AudioCodec) OptionFunc {
	return func(o *Option) { o.AudioCodec = codec }
}

func WithOutputFormat(format OutputFormat) OptionFunc {
	return func(o *Option) { o.OutputFormat = format }
}

// WithVideoBitrate 设置视频目标码率
// 注意：当 CRF > 0 时，CRF 模式优先，此参数仅作为参考
func WithVideoBitrate(bitrate string) OptionFunc {
	return func(o *Option) { o.VideoBitrate = bitrate }
}

func WithAudioBitrate(bitrate string) OptionFunc {
	return func(o *Option) { o.AudioBitrate = bitrate }
}

// WithMaxHeight 设置最大高度，视频会按比例缩放，不会拉伸
func WithMaxHeight(height int) OptionFunc {
	return func(o *Option) { o.MaxHeight = height }
}

// WithMaxWidth 设置最大宽度，视频会按比例缩放，不会拉伸
func WithMaxWidth(width int) OptionFunc {
	return func(o *Option) { o.MaxWidth = width }
}

// WithPreset 设置编码速度预设
// ultrafast > superfast > veryfast > faster > fast > medium > slow > slower > veryslow
// 越慢压缩效率越高，文件越小，但编码时间越长
func WithPreset(preset string) OptionFunc {
	return func(o *Option) { o.Preset = preset }
}

// WithCRF 设置恒定质量因子 (Constant Rate Factor)
// 范围 0-51，值越小质量越高，码率也越高
// 推荐值：18 (接近无损) / 23 (默认) / 28 (较低质量)
// 使用 CRF 时码率是自适应的，会根据画面复杂度自动调整
func WithCRF(crf int) OptionFunc {
	return func(o *Option) { o.CRF = crf }
}

func WithFramerate(fps int) OptionFunc {
	return func(o *Option) { o.Framerate = fps }
}

// WithExtraArgs 添加额外的 ffmpeg 参数
// 例如：WithExtraArgs("-maxrate", "2000k", "-bufsize", "4000k") 设置 VBV 限制
func WithExtraArgs(args ...string) OptionFunc {
	return func(o *Option) { o.ExtraArgs = append(o.ExtraArgs, args...) }
}

// defaultOption 返回默认配置
// 默认使用 CRF 23 质量模式，自适应码率
func defaultOption() *Option {
	return &Option{
		VideoCodec:   VideoCodecH264, // H.264 兼容性最好
		AudioCodec:   AudioCodecAAC,  // AAC 音频
		OutputFormat: OutputFormatMP4,
		AudioBitrate: "128k",
		Preset:       "medium", // 速度与压缩率平衡
		CRF:          23,       // 默认质量，自适应码率
	}
}

func applyOptions(opts ...OptionFunc) *Option {
	o := defaultOption()
	for _, fn := range opts {
		fn(o)
	}
	return o
}
