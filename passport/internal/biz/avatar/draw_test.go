package avatar

import (
	"image/png"
	"math/rand"
	"os"
	"testing"

	xrand "github.com/ryanreadbooks/whimer/misc/xstring/rand"
)

func TestDrawAvatar(t *testing.T) {
	uid := int64(rand.Int63())
	nickname := xrand.Random(18)

	img := DrawImage(uid, nickname)
	f, err := os.OpenFile("tmp.png", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

// BenchmarkDrawAvatarImage 测试DrawAvatarImage函数的性能
func BenchmarkDrawAvatarImage(b *testing.B) {
	// 准备测试数据
	uid := int64(123456789)
	nickname := "测试用户"

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 每次迭代都调用被测函数
		DrawImage(uid, nickname)
	}
}

// BenchmarkDrawAvatarImage_EmptyNickname 测试空昵称情况下DrawAvatarImage函数的性能
func BenchmarkDrawAvatarImage_EmptyNickname(b *testing.B) {
			uid := int64(123456789)
	nickname := ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawImage(uid, nickname)
	}
}

// BenchmarkDrawAvatarImage_LongNickname 测试长昵称情况下DrawAvatarImage函数的性能
func BenchmarkDrawAvatarImage_LongNickname(b *testing.B) {
			uid := int64(123456789)
	nickname := "这是一个非常长的昵称用于测试DrawAvatarImage函数在处理长字符串时的性能表现"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawImage( uid, nickname)
	}
}

// BenchmarkDrawAvatarImage_Parallel 测试DrawAvatarImage函数在并发情况下的性能
func BenchmarkDrawAvatarImage_Parallel(b *testing.B) {
	// 运行并发基准测试
	b.RunParallel(func(pb *testing.PB) {
		// 每个goroutine都有自己的uid和nickname，以避免数据竞争
		uid := int64(123456789)
		nickname := "并发测试用户"

		for pb.Next() {
			DrawImage(uid, nickname)
			// 每次迭代后稍微修改uid，模拟不同用户
			uid++
		}
	})
}
