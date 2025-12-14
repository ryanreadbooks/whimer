package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

type OutputMode int

const (
	OutputModeFile   OutputMode = iota // 落盘模式
	OutputModeStream                   // 流式模式
)

type FFmpeg struct {
	binPath  string
	tempDir  string
	stderrTo io.Writer
}

func New(opts ...func(*FFmpeg)) *FFmpeg {
	f := &FFmpeg{
		binPath: "ffmpeg",
		tempDir: os.TempDir(),
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func WithBinPath(path string) func(*FFmpeg) {
	return func(f *FFmpeg) { f.binPath = path }
}

func WithTempDir(dir string) func(*FFmpeg) {
	return func(f *FFmpeg) { f.tempDir = dir }
}

func WithStderr(w io.Writer) func(*FFmpeg) {
	return func(f *FFmpeg) { f.stderrTo = w }
}

type TranscodeInput struct {
	InputURL string
	Option   *Option
}

type TranscodeOutput struct {
	Reader   io.ReadCloser // 流式模式使用
	FilePath string        // 落盘模式使用
	Cleanup  func()        // 清理临时文件
}

func (f *FFmpeg) Transcode(ctx context.Context, input TranscodeInput, mode OutputMode) (*TranscodeOutput, error) {
	if mode == OutputModeStream {
		return f.transcodeStream(ctx, input)
	}
	return f.transcodeFile(ctx, input)
}

func (f *FFmpeg) transcodeFile(ctx context.Context, input TranscodeInput) (*TranscodeOutput, error) {
	tmpFile := filepath.Join(f.tempDir, uuid.New().String()+".mp4")
	args := f.buildArgs(input.InputURL, input.Option, tmpFile)

	cmd := exec.CommandContext(ctx, f.binPath, args...)
	cmd.Stderr = f.stderrTo

	if err := cmd.Run(); err != nil {
		os.Remove(tmpFile)
		return nil, fmt.Errorf("ffmpeg transcode failed: %w", err)
	}

	return &TranscodeOutput{
		FilePath: tmpFile,
		Cleanup:  func() { os.Remove(tmpFile) },
	}, nil
}

func (f *FFmpeg) transcodeStream(ctx context.Context, input TranscodeInput) (*TranscodeOutput, error) {
	args := f.buildStreamArgs(input.InputURL, input.Option)

	cmd := exec.CommandContext(ctx, f.binPath, args...)
	cmd.Stderr = f.stderrTo

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg stdout pipe failed: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ffmpeg start failed: %w", err)
	}

	return &TranscodeOutput{
		Reader: &cmdReader{ReadCloser: stdout, cmd: cmd},
	}, nil
}

func (f *FFmpeg) buildArgs(inputURL string, opt *Option, output string) []string {
	args := []string{"-y", "-i", inputURL}
	args = append(args, f.buildCommonArgs(opt)...)
	args = append(args, "-movflags", "+faststart")
	args = append(args, output)
	return args
}

func (f *FFmpeg) buildStreamArgs(inputURL string, opt *Option) []string {
	args := []string{"-y", "-i", inputURL}
	args = append(args, f.buildCommonArgs(opt)...)
	args = append(args, "-movflags", "frag_keyframe+empty_moov+default_base_moof")
	args = append(args, "-f", string(opt.OutputFormat), "pipe:1")
	return args
}

func (f *FFmpeg) buildCommonArgs(opt *Option) []string {
	var args []string

	// video
	args = append(args, "-c:v", string(opt.VideoCodec))
	if opt.VideoCodec != VideoCodecCopy {
		if opt.Preset != "" {
			args = append(args, "-preset", opt.Preset)
		}
		if opt.CRF > 0 {
			args = append(args, "-crf", fmt.Sprintf("%d", opt.CRF))
		}
		if opt.VideoBitrate != "" {
			args = append(args, "-b:v", opt.VideoBitrate)
		}
		if opt.MaxHeight > 0 || opt.MaxWidth > 0 {
			args = append(args, "-vf", f.buildScaleFilter(opt))
		}
		if opt.Framerate > 0 {
			args = append(args, "-r", fmt.Sprintf("%d", opt.Framerate))
		}
	}

	// audio
	args = append(args, "-c:a", string(opt.AudioCodec))
	if opt.AudioCodec != AudioCodecCopy && opt.AudioBitrate != "" {
		args = append(args, "-b:a", opt.AudioBitrate)
	}

	args = append(args, opt.ExtraArgs...)
	return args
}

// buildScaleFilter 构建缩放滤镜
// 同时兼容横屏和竖屏：
//   - 横屏 (宽>高): 宽度超过限制时按宽度缩放
//   - 竖屏 (高>宽): 高度超过限制时按高度缩放
//   - 保持宽高比，不拉伸
//   - 不放大，只缩小
//   - 自动偶数对齐
func (f *FFmpeg) buildScaleFilter(opt *Option) string {
	maxDim := opt.MaxHeight
	if opt.MaxWidth > 0 && (opt.MaxHeight == 0 || opt.MaxWidth < opt.MaxHeight) {
		maxDim = opt.MaxWidth
	}

	// 使用 iw/ih 判断横竖屏，取较长边与限制比较
	// 横屏: max(iw,ih)=iw，如果 iw > maxDim 则缩放
	// 竖屏: max(iw,ih)=ih，如果 ih > maxDim 则缩放
	// force_original_aspect_ratio=decrease 保证只缩小不放大，保持宽高比
	filter := fmt.Sprintf("scale='if(gte(iw,ih),min(%d,iw),-2)':'if(gte(iw,ih),-2,min(%d,ih))'", maxDim, maxDim)

	// 偶数对齐，避免编码器报错
	filter += ",pad=ceil(iw/2)*2:ceil(ih/2)*2"

	return filter
}

func (f *FFmpeg) ExtractThumbnail(ctx context.Context, inputURL string, atSecond float64) (string, func(), error) {
	tmpFile := filepath.Join(f.tempDir, uuid.New().String()+".jpg")
	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.2f", atSecond),
		"-i", inputURL,
		"-vframes", "1",
		"-q:v", "2",
		tmpFile,
	}

	cmd := exec.CommandContext(ctx, f.binPath, args...)
	cmd.Stderr = f.stderrTo

	if err := cmd.Run(); err != nil {
		os.Remove(tmpFile)
		return "", nil, fmt.Errorf("ffmpeg extract thumbnail failed: %w", err)
	}

	return tmpFile, func() { os.Remove(tmpFile) }, nil
}

type cmdReader struct {
	io.ReadCloser
	cmd *exec.Cmd
}

func (r *cmdReader) Close() error {
	r.ReadCloser.Close()
	return r.cmd.Wait()
}
