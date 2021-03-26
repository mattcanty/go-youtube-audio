package mp3

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/matt.canty/go-youtube-audio/pkg/models"
	"github.com/matt.canty/go-youtube-audio/pkg/youtube"
	"github.com/xfrr/goffmpeg/transcoder"
)

func Download(videoID string, outputDirectory string) error {
	if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
		return err
	}

	info, err := youtube.GetVideoInfo(videoID)
	if err != nil {
		return err
	}

	var selectedFormat models.AdaptiveFormat
	for _, format := range info.PlayerResponse.StreamingData.AdaptiveFormats {
		if !strings.Contains(format.MimeType, "audio/webm") {
			continue
		}

		selectedFormat = format
		break
	}

	webmFileName := fmt.Sprintf("%s.webm", info.PlayerResponse.VideoDetails.Title)
	mp3FileName := fmt.Sprintf("%s.mp3", info.PlayerResponse.VideoDetails.Title)

	webmPath := path.Join(os.TempDir(), webmFileName)
	mp3Path := path.Join(os.TempDir(), mp3FileName)

	webmFile, err := os.Create(webmPath)
	mp3File, err := os.Create(mp3Path)

	err = youtube.DownloadAudio(selectedFormat.URL, webmFile)
	if err != nil {
		return err
	}

	trans := new(transcoder.Transcoder)
	err = trans.Initialize(webmFile.Name(), mp3File.Name())
	if err != nil {
		return err
	}

	done := trans.Run(true)

	err = <-done

	os.Remove(webmPath)
	os.Rename(mp3Path, path.Join(outputDirectory, mp3FileName))

	return err
}
