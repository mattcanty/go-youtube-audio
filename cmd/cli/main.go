package main

import (
	"os"

	"github.com/matt.canty/go-youtube-audio/pkg/mp3"
)

func main() {
	mp3.Download(os.Args[1], os.Args[2])
}
