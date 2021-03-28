package mp3

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kennygrant/sanitize"
	"github.com/matt.canty/go-youtube-audio/internal/logger"
	"github.com/matt.canty/go-youtube-audio/internal/youtube"
	"github.com/matt.canty/go-youtube-audio/pkg/models"
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

		logger.Debug(fmt.Sprintf("Selected '%s' %s ", format.MimeType, format.AudioQuality))

		selectedFormat = format
		break
	}

	santitisedTitle := sanitize.Path(info.PlayerResponse.VideoDetails.Title)

	webmFileName := fmt.Sprintf("%s.webm", santitisedTitle)
	mp3FileName := fmt.Sprintf("%s.mp3", santitisedTitle)

	webmPath := filepath.FromSlash(path.Join(os.TempDir(), webmFileName))
	mp3Path := filepath.FromSlash(path.Join(os.TempDir(), mp3FileName))

	webmFile, err := os.Create(webmPath)
	if err != nil {
		return err
	}

	mp3File, err := os.Create(mp3Path)
	if err != nil {
		return err
	}

	err = youtube.DownloadAudio(selectedFormat.URL, webmFile)
	if err != nil {
		return err
	}

	logger.Debug(fmt.Sprintf("Writing audio to: '%s'", webmPath))

	trans := new(transcoder.Transcoder)
	err = trans.Initialize(webmFile.Name(), mp3File.Name())
	if err != nil {
		return err
	}

	done := trans.Run(true)

	err = <-done

	os.Remove(webmPath)
	os.Rename(mp3Path, filepath.FromSlash(path.Join(outputDirectory, mp3FileName)))

	return err
}
