package mp3

import (
	"github.com/mattcanty/go-ffmpeg-transcode/pkg/ffmpeg"
)

func transcode(webmFileName string, mp3FileName string) error {
	return ffmpeg.Transcode(webmFileName, mp3FileName)
}
