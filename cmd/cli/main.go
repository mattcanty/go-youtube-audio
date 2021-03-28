package main

import (
	"os"
	"path/filepath"

	"github.com/matt.canty/go-youtube-audio/internal/logger"
	"github.com/matt.canty/go-youtube-audio/internal/mp3"
)

func main() {
	videoID := os.Args[1]
	downloadDirectory, err := filepath.Abs(os.Args[2])

	if err != nil {
		logger.Fatal(logger.InvalidDownloadDirectory, err)
	}

	logger.Info(logger.DownloadingVideo)

	err = mp3.Download(videoID, downloadDirectory)
	if err != nil {
		logger.Fatal(logger.FailedToDownloadMP3, err)
	}

	logger.Info(logger.FinishedDownloadingVideo)
}
