package mp3

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/matt.canty/go-youtube-audio/internal/logger"
	"github.com/xfrr/goffmpeg/ffmpeg"
	"github.com/xfrr/goffmpeg/transcoder"
)

func transcode(webmFileName string, mp3FileName string) error {
	logger.Debug(fmt.Sprint("Windows OS detected"))

	trans := new(transcoder.Transcoder)

	if runtime.GOOS == "windows" {
		conf, _ := ffmpeg.Configure()

		ffprobeBin := strings.SplitAfterN(conf.FfprobeBin, "ffprobe.exe", 2)[0]
		ffmpegBin := strings.SplitAfterN(conf.FfmpegBin, "ffmpeg.exe", 2)[0]

		logger.Debug(fmt.Sprintf("ffprobe bin path fixed to %s", ffprobeBin))
		logger.Debug(fmt.Sprintf("ffmpeg bin path fixed to %s", ffmpegBin))

		conf.FfprobeBin = ffprobeBin
		conf.FfmpegBin = ffmpegBin

		trans.SetConfiguration(conf)
	}

	err := trans.Initialize(webmFileName, mp3FileName)
	if err != nil {
		return err
	}

	done := trans.Run(true)

	err = <-done

	return err
}
