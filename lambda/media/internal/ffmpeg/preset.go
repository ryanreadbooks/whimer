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

// 标准分辨率定义（长边像素）
// 这些是业界通用的标准分辨率，兼容性最好
const (
	Resolution4K    = 3840 // 4K/2160p
	Resolution1440p = 2560 // 1440p/2K
	Resolution1080p = 1920 // 1080p/FHD
	Resolution720p  = 1280 // 720p/HD
	Resolution480p  = 854  // 480p/SD
	Resolution360p  = 640  // 360p
)

// 标准分辨率列表（降序排列）
var StandardResolutions = []int{
	Resolution4K,
	Resolution1440p,
	Resolution1080p,
	Resolution720p,
	Resolution480p,
	Resolution360p,
}

// IsStandardResolution 判断是否为标准分辨率
// 允许 ±2% 的误差（考虑到一些设备/编码器的微小差异）
func IsStandardResolution(width, height int) bool {
	maxDim := width
	if height > width {
		maxDim = height
	}

	const tolerance = 0.02 // 2% 误差容忍度
	for _, std := range StandardResolutions {
		minVal := int(float64(std) * (1 - tolerance))
		maxVal := int(float64(std) * (1 + tolerance))
		if maxDim >= minVal && maxDim <= maxVal {
			return true
		}
	}
	return false
}

// RecommendedScaleTarget 根据源分辨率推荐缩放目标
// 返回值：
//   - targetMaxDim: 推荐缩放到的长边像素，0 表示不需要缩放
//   - needsScale: 是否需要缩放
//
// 策略：
//   - 标准分辨率（1080p/720p/480p等）：不处理，保持原样
//   - 超过 1080p 的非标准分辨率：缩放到 1080p
//   - 720p ~ 1080p 之间的非标准分辨率：缩放到 720p（兼容性最好）
//   - 480p ~ 720p 之间的非标准分辨率：缩放到 720p
//   - 小于 480p：不处理
func RecommendedScaleTarget(width, height int) (targetMaxDim int, needsScale bool) {
	// 标准分辨率不处理
	if IsStandardResolution(width, height) {
		return 0, false
	}

	maxDim := width
	if height > width {
		maxDim = height
	}

	// 根据当前分辨率决定缩放目标
	switch {
	case maxDim > Resolution1080p:
		// 超过 1080p 的非标准分辨率，缩放到 1080p
		return Resolution1080p, true
	case maxDim > Resolution720p:
		// 720p ~ 1080p 之间的非标准分辨率（如 1600x900, 1440x810）
		// 缩放到 720p，因为 720p 兼容性最好，且避免产生非标准输出
		return Resolution720p, true
	case maxDim > Resolution480p:
		// 480p ~ 720p 之间的非标准分辨率（如 960x540）
		// 缩放到 720p（向上取整到标准分辨率会放大，所以保持或向下）
		// 这里选择不缩放，保持原样，避免质量损失
		return 0, false
	default:
		// 小于等于 480p，不处理
		return 0, false
	}
}

// NeedsScaleTo720p 判断是否需要缩放到 720p（兼容旧接口）
func NeedsScaleTo720p(srcWidth, srcHeight int) bool {
	target, needs := RecommendedScaleTarget(srcWidth, srcHeight)
	return needs && target == Resolution720p
}

// AV1 编码推荐参数
// Preset: 0-13，越小越慢质量越好，推荐 6-8 平衡速度与质量
// CRF: 0-63，越小质量越高，推荐 30-35
const (
	AV1PresetFast   = "10" // 快速编码
	AV1PresetMedium = "8"  // 平衡
	AV1PresetSlow   = "6"  // 高质量

	AV1CRFHigh   = 28 // 高质量
	AV1CRFMedium = 35 // 默认质量
	AV1CRFLow    = 42 // 低质量/小文件
)

// WithAV1Default 应用 AV1 默认编码参数
func WithAV1Default() OptionFunc {
	return func(o *Option) {
		o.VideoCodec = VideoCodecAV1
		o.Preset = AV1PresetMedium
		o.CRF = AV1CRFMedium
	}
}

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
