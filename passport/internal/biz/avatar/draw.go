package avatar

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

func EncodeToPng(img image.Image) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := png.Encode(buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DrawImage 生成基于用户ID和昵称的头像图像
func DrawImage(uid int64, nickname string) image.Image {
	const (
		size      = 180              // 图像尺寸
		numBlocks = 6                // 6x6 色块
		blockSize = size / numBlocks // 每个色块的大小
	)

	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	bgColor := generateRandomBgColor(uid)
	// 填充背景色
	draw.Draw(img, img.Bounds(), image.NewUniform(bgColor), image.Point{}, draw.Src)

	binaryPattern := encodeNicknameToBitPattern(nickname, numBlocks*numBlocks)

	for y := 0; y < numBlocks; y++ {
		for x := 0; x < numBlocks; x++ {
			// 计算当前位置在数据中的索引
			index := y*numBlocks + x
			// 检查该位置是否需要绘制色块
			if (binaryPattern & (1 << index)) != 0 {
				// 计算色块位置
				x0 := x * blockSize
				y0 := y * blockSize
				x1 := x0 + blockSize
				y1 := y0 + blockSize

				// 确保边界不超出图像
				if x1 > size {
					x1 = size
				}
				if y1 > size {
					y1 = size
				}

				// 根据背景色动态调整色块颜色
				blockColor := generateBlockColor(nickname, x, y, bgColor)

				// 绘制色块
				draw.Draw(img, image.Rect(x0, y0, x1, y1),
					image.NewUniform(blockColor), image.Point{}, draw.Src)
			}
		}
	}

	return img
}

// 将昵称编码为指定位数的二进制数据
func encodeNicknameToBitPattern(nickname string, bits int) uint64 {
	if len(nickname) == 0 {
		return 0
	}

	// 计算昵称的哈希值
	hash := uint64(0)
	for i := 0; i < len(nickname); i++ {
		// 使用类似fnv-1a的哈希算法
		hash ^= uint64(nickname[i])
		hash *= 1099511628211
	}

	mask := uint64(0)
	for i := range bits {
		mask |= (1 << i)
	}

	return hash & mask
}

// 基于昵称、位置和背景色动态生成色块颜色
func generateBlockColor(nickname string, x, y int, bgColor color.Color) color.Color {
	// 计算昵称的哈希值
	hash := uint64(0)
	for i := 0; i < len(nickname); i++ {
		hash ^= uint64(nickname[i])
		hash *= 1099511628211
	}

	// 将位置信息加入哈希
	hash ^= uint64(x) << 32
	hash ^= uint64(y) << 40

	// 提取背景色的RGB值
	bgR, bgG, bgB, _ := bgColor.RGBA()
	bgR = bgR >> 8
	bgG = bgG >> 8
	bgB = bgB >> 8

	r, g, b := uint8((hash>>0)%100+100), uint8((hash>>8)%100+100), uint8((hash>>16)%100+100)

	// 根据位置微调颜色，增强区分度
	r = adjustColorByPosition(r, x, y)
	g = adjustColorByPosition(g, x, y)
	b = adjustColorByPosition(b, x, y)

	// 确保颜色与背景有足够对比度
	r = ensureContrast(r, bgR)
	g = ensureContrast(g, bgG)
	b = ensureContrast(b, bgB)

	return color.RGBA{r, g, b, 255}
}

// 根据位置微调颜色值
func adjustColorByPosition(colorVal uint8, x, y int) uint8 {
	positionFactor := int(x*3+y*5) % 20
	adjustedVal := int(colorVal) - positionFactor/2
	if adjustedVal < 30 {
		adjustedVal = 30
	} else if adjustedVal > 220 {
		adjustedVal = 220
	}
	return uint8(adjustedVal)
}

// 确保颜色与背景有足够对比度
func ensureContrast(colorVal uint8, bgColorVal uint32) uint8 {
	// 计算颜色差异
	diff := int(colorVal) - int(bgColorVal)
	if diff < 0 {
		diff = -diff
	}

	// 如果对比度不够，调整颜色
	if diff < 60 {
		// 根据背景色深浅调整
		if bgColorVal > 160 {
			// 背景较亮，调暗色
			if int(colorVal) > 100 {
				colorVal = uint8(int(colorVal) - 60)
			}
		} else {
			// 背景较暗，调亮色
			if int(colorVal) < 160 {
				colorVal = uint8(int(colorVal) + 60)
			}
		}
	}

	return colorVal
}

// 生成基于用户ID的随机背景色
func generateRandomBgColor(uid int64) color.Color {
	return color.RGBA{
		R: uint8((uid>>0)%55 + 200),
		G: uint8((uid>>8)%55 + 200),
		B: uint8((uid>>16)%55 + 200),
		A: 255,
	}
}
