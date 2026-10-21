package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "make-audio":
		if err := makeAudio(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`vidtune - turn music videos into audio files with cover art

Usage:
  vidtune make-audio    convert every video in the current folder`)
}
func checkDeps() error {
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s not found in PATH, install it with: sudo dnf install ffmpeg-free", tool)
		}
	}
	return nil
}
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
func makeAudio() error {
	if err := checkDeps(); err != nil {
		return err
	}

	videos, err := findVideos(".")
	if err != nil {
		return err
	}
	if len(videos) == 0 {
		return fmt.Errorf("no video files found in this folder")
	}

	fmt.Printf("Found %d video(s):\n", len(videos))
	for _, v := range videos {
		fmt.Println("  -", v)
	}
	return nil
}