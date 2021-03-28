package youtube

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/matt.canty/go-youtube-audio/internal/logger"
	"github.com/matt.canty/go-youtube-audio/pkg/models"
	"github.com/schollz/progressbar/v3"
)

func GetVideoInfo(videoID string) (*models.VideoInfo, error) {
	videoInfoURL := fmt.Sprintf("https://www.youtube.com/get_video_info?video_id=%s", videoID)
	logger.Debug(fmt.Sprintf("Getting video info from URL: '%s'", videoInfoURL))

	var client http.Client
	resp, err := client.Get(videoInfoURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	videoInfo := &models.VideoInfo{}
	err = models.Decode(string(bodyBytes), videoInfo)
	if err != nil {
		return nil, err
	}

	return videoInfo, err
}

func DownloadAudio(audioURL string, file *os.File, expectedBytes int64) error {
	logger.Debug(fmt.Sprintf("Downloading audio from URL: '%s' to file: '%s'", audioURL, file.Name()))
	client := &http.Client{}
	request, err := http.NewRequest("GET", audioURL, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Cache-Control", "public")
	request.Header.Set("Content-Description", "File Transfer")
	request.Header.Set("Content-Disposition", "attachment; filename="+file.Name())
	request.Header.Set("Content-Type", "application/zip")
	request.Header.Set("Content-Transfer-Encoding", "binary")
	request.Header.Set("Range", fmt.Sprintf("bytes=0-%d", expectedBytes))

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("%s: %s", http.StatusText(response.StatusCode), response.Status)
	}

	counter := &WriteCounter{
		Bar: progressbar.DefaultBytes(expectedBytes),
	}
	if _, err = io.Copy(file, io.TeeReader(response.Body, counter)); err != nil {
		return err
	}

	logger.Debug(fmt.Sprintf("Writing audio to: '%s'", file.Name()))
	_, err = io.Copy(file, response.Body)

	return nil
}

type WriteCounter struct {
	Bar *progressbar.ProgressBar
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Bar.Add(n)

	return n, nil
}
