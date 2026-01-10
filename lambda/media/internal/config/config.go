package config

import (
	"fmt"
	"os/exec"

	"github.com/ryanreadbooks/whimer/lambda/media/internal/storage"
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/zeromicro/go-zero/core/logx"
)

var Conf Config

type Config struct {
	Log logx.LogConf `json:"log"`

	Conductor xconf.Discovery `json:"conductor"`

	Worker struct {
		Concurrency int `json:"concurrency"`
	} `json:"worker"`

	Storage storage.Config `json:"storage"`

	FFmpeg struct {
		BinPath string `json:"binPath"`
		TempDir string `json:"tempDir"`
	} `json:"ffmpeg"`
}

func MustInit() {
	path, err := exec.LookPath(Conf.FFmpeg.BinPath)
	if err != nil {
		panic(fmt.Errorf("ffmpeg bin path %s not found or not executable: %w", Conf.FFmpeg.BinPath, err))
	}
	Conf.FFmpeg.BinPath = path
}
