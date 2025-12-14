package ffmpeg

// Preset 转码预设配置
type Preset struct {
	Name         string // 预设名称，如 "1080p", "720p"
	MaxHeight    int    // 最长边限制（像素），横屏限制宽度，竖屏限制高度
	VideoBitrate string // 参考码率，CRF 模式下仅作参考
	AudioBitrate string // 音频码率
}

// 预定义的转码档位
// 码率为参考值，实际使用 CRF 模式时会自适应调整
var (
	Preset1080p = Preset{Name: "1080p", MaxHeight: 1920, VideoBitrate: "2500k", AudioBitrate: "128k"}
	Preset720p  = Preset{Name: "720p", MaxHeight: 1280, VideoBitrate: "1500k", AudioBitrate: "128k"}

	DefaultPresets = []Preset{Preset1080p, Preset720p}
)

// ToOptions 将预设转换为 OptionFunc 列表
func (p Preset) ToOptions() []OptionFunc {
	return []OptionFunc{
		WithMaxHeight(p.MaxHeight),
		WithVideoBitrate(p.VideoBitrate),
		WithAudioBitrate(p.AudioBitrate),
	}
}

// SelectPresets 根据源视频分辨率选择合适的转码档位
// 只返回不超过源分辨率的档位，避免放大
func SelectPresets(srcWidth, srcHeight int) []Preset {
	maxDim := srcWidth
	if srcHeight > srcWidth {
		maxDim = srcHeight
	}

	var result []Preset
	for _, p := range DefaultPresets {
		if p.MaxHeight <= maxDim {
			result = append(result, p)
		}
	}

	// 至少保留一个最低档位
	if len(result) == 0 && len(DefaultPresets) > 0 {
		result = append(result, DefaultPresets[len(DefaultPresets)-1])
	}

	return result
}
