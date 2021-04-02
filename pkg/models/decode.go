package models

import (
	"encoding/json"
	"net/url"
	"strings"
)

func Decode(videoInfoData string, videoInfo *VideoInfo) error {
	for _, item := range strings.Split(videoInfoData, "&") {
		kv := strings.SplitN(item, "=", 2)
		switch kv[0] {
		case "player_response":
			jsonData, err := url.QueryUnescape(kv[1])
			if err != nil {
				return err
			}

			err = json.Unmarshal([]byte(jsonData), &videoInfo.PlayerResponse)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
