package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func checkDeps() error {
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s not found in PATH, install it with: sudo dnf install ffmpeg-free", tool)
		}
	}
	return nil
}

// runFFmpeg runs ffmpeg quietly, at the lowest CPU priority, and returns
// its own error text on failure.
func runFFmpeg(args ...string) error {
	ffArgs := append([]string{"-y", "-loglevel", "error", "-filter_threads", "1"}, args...)

	name, full := "ffmpeg", ffArgs
	// "nice -n 19" makes the OS give ffmpeg CPU time only when nothing
	// else needs it. If nice is missing, we just run ffmpeg normally.
	if nicePath, err := exec.LookPath("nice"); err == nil {
		name = nicePath
		full = append([]string{"-n", "19", "ffmpeg"}, ffArgs...)
	}

	cmd := exec.Command(name, full...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// videoDuration asks ffprobe for the video length in seconds.
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

// audioCodec returns the codec name of the first audio stream.
func audioCodec(path string) (string, error) {
	out, err := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	).Output()
	if err != nil {
		return "", fmt.Errorf("ffprobe failed for %s: %w", path, err)
	}

	codec := strings.TrimSpace(string(out))
	if codec == "" {
		return "", fmt.Errorf("%s has no audio stream", path)
	}
	return codec, nil
}