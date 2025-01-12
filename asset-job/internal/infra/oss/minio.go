package oss

import (
	"github.com/ryanreadbooks/whimer/asset-job/internal/config"
	"github.com/ryanreadbooks/whimer/misc/oss/downloader"
	"github.com/ryanreadbooks/whimer/misc/oss/uploader"
)

var (
	upload   *uploader.Uploader
	download *downloader.Downloader
)

func Init(c *config.Config) {
	var err error
	upload, err = uploader.New(uploader.Config{
		Ak:       c.NoteOss.Ak,
		Sk:       c.NoteOss.Sk,
		Endpoint: c.NoteOss.Endpoint,
		Location: c.NoteOss.Location,
	})
	if err != nil {
		panic(err)
	}

	download, err = downloader.New(downloader.Config{
		Ak:       c.NoteOss.Ak,
		Sk:       c.NoteOss.Sk,
		Endpoint: c.NoteOss.Endpoint,
		Location: c.NoteOss.Location,
	})
	if err != nil {
		panic(err)
	}
}

func Uploader() *uploader.Uploader {
	return upload
}

func Downloader() *downloader.Downloader {
	return download
}
