package main

import (
	"os"
	"path/filepath"
	"strings"
)

var videoExts = map[string]bool{
	".mp4": true, ".mkv": true, ".webm": true,
	".mov": true, ".avi": true, ".m4v": true,
}

func findVideos(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var videos []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if videoExts[ext] {
			videos = append(videos, e.Name())
		}
	}
	return videos, nil
}