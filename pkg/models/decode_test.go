package models

import (
	"io/ioutil"
	"testing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestDecode(t *testing.T) {
	sample, err := ioutil.ReadFile("sample")
	check(err)

	var videoInfo VideoInfo
	err = Decode(string(sample), &videoInfo)
	check(err)

	if videoInfo.PlayerResponse.VideoDetails.Title != "Zombie Kid Likes Turtles" {
		t.Errorf("Unable to get video title from info.")
	}
}
