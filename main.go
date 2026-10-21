package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// NEW: ask ffprobe for the video length in seconds.
func videoDuration(path string) (float64, error) {
	out, err := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	).Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed for %s: %w", path, err)
	}

	secs, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("could not read duration of %s: %w", path, err)
	}
	return secs, nil
}

// NEW: choose the moment to screenshot.
// Music videos often start with black frames or logos and end with credits,
// so we jump 30% in, where the performance usually is.
func pickTimestamp(duration float64) float64 {
	if duration < 10 {
		return duration / 2 // very short clip: just take the middle
	}
	return duration * 0.30
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
		dur, err := videoDuration(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, "  skipping:", err)
			continue
		}
		ts := pickTimestamp(dur)
		fmt.Printf("  - %s (%.0fs long, screenshot at %.1fs)\n", v, dur, ts)
	}
	return nil
}