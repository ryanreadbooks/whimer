package ffmpeg_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/lambda/media/internal/ffmpeg"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/storage"
)

const (
	testInputURL = ""
	testBucket   = "devvideo"
)

func getTestStorage(t *testing.T) *storage.Storage {
	store, err := storage.New(storage.Config{
		Endpoint: getEnvOrDefault("ENV_OSS_ENDPOINT", "localhost:9000"),
		Ak:       getEnvOrDefault("ENV_OSS_AK", "minioadmin"),
		Sk:       getEnvOrDefault("ENV_OSS_SK", "minioadmin"),
		UseSSL:   false,
	})
	if err != nil {
		t.Fatalf("create storage failed: %v", err)
	}
	return store
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func TestProcessorStreamH264(t *testing.T) {
	store := getTestStorage(t)
	ff := ffmpeg.New(ffmpeg.WithStderr(os.Stderr))
	processor := ffmpeg.NewProcessor(ff, store)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := processor.ProcessSingle(ctx, ffmpeg.SingleProcessRequest{
		InputURL:  testInputURL,
		Bucket:    testBucket,
		OutputKey: "test/stream_h264.mp4",
		Options: []ffmpeg.OptionFunc{
			ffmpeg.WithVideoCodec(ffmpeg.VideoCodecH264),
			ffmpeg.WithCRF(23),
			ffmpeg.WithPreset("fast"),
			ffmpeg.WithMaxHeight(720),
		},
	})
	if err != nil {
		t.Fatalf("process h264 failed: %v", err)
	}

	t.Logf("H.264 result: %+v", result)
}

func TestProcessorStreamH265(t *testing.T) {
	store := getTestStorage(t)
	ff := ffmpeg.New(ffmpeg.WithStderr(os.Stderr))
	processor := ffmpeg.NewProcessor(ff, store)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := processor.ProcessSingle(ctx, ffmpeg.SingleProcessRequest{
		InputURL:  testInputURL,
		Bucket:    testBucket,
		OutputKey: "test/stream_h265.mp4",
		Options: []ffmpeg.OptionFunc{
			ffmpeg.WithVideoCodec(ffmpeg.VideoCodecH265),
			ffmpeg.WithCRF(28),
			ffmpeg.WithPreset("fast"),
			ffmpeg.WithMaxHeight(720),
		},
	})
	if err != nil {
		t.Fatalf("process h265 failed: %v", err)
	}

	t.Logf("H.265 result: %+v", result)
}

func TestProcessorStreamAV1(t *testing.T) {
	store := getTestStorage(t)
	ff := ffmpeg.New(ffmpeg.WithStderr(os.Stderr))
	processor := ffmpeg.NewProcessor(ff, store)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := processor.ProcessSingle(ctx, ffmpeg.SingleProcessRequest{
		InputURL:  testInputURL,
		Bucket:    testBucket,
		OutputKey: "test/stream_av1.mp4",
		Options: []ffmpeg.OptionFunc{
			ffmpeg.WithVideoCodec(ffmpeg.VideoCodecAV1),
			ffmpeg.WithCRF(35),
			ffmpeg.WithPreset("10"), // SVT-AV1 preset 0-13
			ffmpeg.WithMaxHeight(720),
		},
	})
	if err != nil {
		t.Fatalf("process av1 failed: %v", err)
	}

	t.Logf("AV1 result: %+v", result)
}

func TestProcessorStreamAll(t *testing.T) {
	store := getTestStorage(t)
	ff := ffmpeg.New(ffmpeg.WithStderr(os.Stderr))
	processor := ffmpeg.NewProcessor(ff, store)

	tests := []struct {
		name      string
		outputKey string
		options   []ffmpeg.OptionFunc
		timeout   time.Duration
	}{
		{
			name:      "H.264",
			outputKey: "test/stream_h264.mp4",
			options: []ffmpeg.OptionFunc{
				ffmpeg.WithVideoCodec(ffmpeg.VideoCodecH264),
				ffmpeg.WithCRF(23),
				ffmpeg.WithPreset("fast"),
				ffmpeg.WithMaxHeight(720),
			},
			timeout: 5 * time.Minute,
		},
		{
			name:      "H.265",
			outputKey: "test/stream_h265.mp4",
			options: []ffmpeg.OptionFunc{
				ffmpeg.WithVideoCodec(ffmpeg.VideoCodecH265),
				ffmpeg.WithCRF(28),
				ffmpeg.WithPreset("fast"),
				ffmpeg.WithMaxHeight(720),
			},
			timeout: 5 * time.Minute,
		},
		{
			name:      "AV1",
			outputKey: "test/stream_av1.mp4",
			options: []ffmpeg.OptionFunc{
				ffmpeg.WithVideoCodec(ffmpeg.VideoCodecAV1),
				ffmpeg.WithCRF(35),
				ffmpeg.WithPreset("10"),
				ffmpeg.WithMaxHeight(720),
			},
			timeout: 10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			start := time.Now()
			result, err := processor.ProcessSingle(ctx, ffmpeg.SingleProcessRequest{
				InputURL:  testInputURL,
				Bucket:    testBucket,
				OutputKey: tt.outputKey,
				Options:   tt.options,
			})
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("process failed: %v", err)
			}

			t.Logf("%s completed in %v", tt.name, elapsed)
			t.Logf("Result: width=%d, height=%d, duration=%.2fs, bitrate=%d, codec=%s, fps=%.2f",
				result.Width, result.Height, result.Duration, result.Bitrate, result.Codec, result.Framerate)
			t.Logf("Audio: codec=%s, sampleRate=%d, channels=%d, bitrate=%d",
				result.AudioCodec, result.AudioSampleRate, result.AudioChannels, result.AudioBitrate)
		})
	}
}
