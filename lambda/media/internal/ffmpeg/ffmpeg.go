package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
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

// Transcode 转码视频，输出到 stdout 流
// 使用 fragmented MP4 格式，适合流式传输但浏览器渐进式下载支持较差
func (f *FFmpeg) Transcode(ctx context.Context, input TranscodeInput) (io.ReadCloser, error) {
	args := f.buildArgs(input.InputURL, input.Option)

	cmd := exec.CommandContext(ctx, f.binPath, args...)

	// stderr 用于错误诊断
	var stderrBuf bytes.Buffer
	if f.stderrTo != nil {
		cmd.Stderr = io.MultiWriter(f.stderrTo, &stderrBuf)
	} else {
		cmd.Stderr = &stderrBuf
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg stdout pipe failed: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("ffmpeg start failed: %w", err)
	}

	return &cmdReader{ReadCloser: stdout, cmd: cmd, stderr: &stderrBuf}, nil
}

// TranscodeToFile 转码视频并输出到文件
// 使用 faststart 优化，moov atom 在文件头，支持 HTTP 渐进式下载
func (f *FFmpeg) TranscodeToFile(ctx context.Context, inputURL, outputPath string, opt *Option) error {
	args := f.buildArgsForFile(inputURL, outputPath, opt)

	cmd := exec.CommandContext(ctx, f.binPath, args...)

	var stderrBuf bytes.Buffer
	if f.stderrTo != nil {
		cmd.Stderr = io.MultiWriter(f.stderrTo, &stderrBuf)
	} else {
		cmd.Stderr = &stderrBuf
	}

	if err := cmd.Run(); err != nil {
		stderrStr := stderrBuf.String()
		if len(stderrStr) > 500 {
			stderrStr = "..." + stderrStr[len(stderrStr)-500:]
		}
		return fmt.Errorf("ffmpeg transcode to file failed: %w: %s", err, stderrStr)
	}

	return nil
}

// buildArgs 构建流式输出参数（输出到管道）
func (f *FFmpeg) buildArgs(inputURL string, opt *Option) []string {
	args := []string{"-y", "-i", inputURL}
	args = append(args, f.buildCommonArgs(opt)...)
	args = append(args, "-movflags", "frag_keyframe+empty_moov+default_base_moof")
	args = append(args, "-f", string(opt.OutputFormat), "pipe:1")
	return args
}

// buildArgsForFile 构建文件输出参数（使用 faststart 优化渐进式下载）
func (f *FFmpeg) buildArgsForFile(inputURL, outputPath string, opt *Option) []string {
	args := []string{"-y", "-i", inputURL}
	args = append(args, f.buildCommonArgs(opt)...)
	// faststart: 将 moov atom 移到文件开头，支持 HTTP 渐进式下载
	args = append(args, "-movflags", "+faststart")
	args = append(args, "-f", string(opt.OutputFormat), outputPath)
	return args
}

func (f *FFmpeg) buildCommonArgs(opt *Option) []string {
	var args []string

	// video
	args = append(args, "-c:v", string(opt.VideoCodec))
	if opt.VideoCodec != VideoCodecCopy {
		args = append(args, f.buildVideoEncoderArgs(opt)...)
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

// buildVideoEncoderArgs 根据编码器构建特定参数
func (f *FFmpeg) buildVideoEncoderArgs(opt *Option) []string {
	var args []string

	switch opt.VideoCodec {
	case VideoCodecAV1:
		// SVT-AV1 参数
		if opt.Preset != "" {
			// AV1 preset 需要是数字 0-13，转换字符串格式
			preset := convertAV1Preset(opt.Preset)
			args = append(args, "-preset", preset)
		}
		if opt.CRF > 0 {
			args = append(args, "-crf", fmt.Sprintf("%d", opt.CRF)) // 0-63，默认 35
		}
		if opt.VideoBitrate != "" {
			args = append(args, "-b:v", opt.VideoBitrate)
		}
	default:
		// H.264/H.265 参数
		if opt.Preset != "" {
			args = append(args, "-preset", opt.Preset)
		}
		if opt.CRF > 0 {
			args = append(args, "-crf", fmt.Sprintf("%d", opt.CRF))
		}
		if opt.VideoBitrate != "" {
			args = append(args, "-b:v", opt.VideoBitrate)
		}
	}

	return args
}

// convertAV1Preset 将 H.264 风格的 preset 字符串转换为 AV1 数字格式
func convertAV1Preset(preset string) string {
	switch preset {
	case "ultrafast":
		return "12"
	case "superfast", "veryfast":
		return "11"
	case "faster":
		return "10"
	case "fast":
		return "9"
	case "medium":
		return "8"
	case "slow":
		return "6"
	case "slower":
		return "4"
	case "veryslow":
		return "2"
	default:
		return preset // 已经是数字格式，直接返回
	}
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
	cmd    *exec.Cmd
	stderr *bytes.Buffer
}

func (r *cmdReader) Close() error {
	r.ReadCloser.Close()
	err := r.cmd.Wait()
	if err != nil && r.stderr != nil && r.stderr.Len() > 0 {
		// 截取最后 500 字节的错误信息
		stderrStr := r.stderr.String()
		if len(stderrStr) > 500 {
			stderrStr = "..." + stderrStr[len(stderrStr)-500:]
		}
		return fmt.Errorf("%w: %s", err, stderrStr)
	}
	return err
}
