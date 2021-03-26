package youtube

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/matt.canty/go-youtube-audio/pkg/models"
)

func GetVideoInfo(videoID string) (*models.VideoInfo, error) {
	var client http.Client
	resp, err := client.Get("https://www.youtube.com/get_video_info?video_id=" + videoID)
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

func DownloadAudio(audioURL string, file *os.File) error {
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

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", http.StatusText(response.StatusCode), response.Status)
	}

	_, err = io.Copy(file, response.Body)

	return nil
}
