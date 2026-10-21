package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// pickTimestamp chooses the moment to screenshot. Music videos often start
// with black frames or logos, so we jump 30% in, where the performance
// usually is.
func pickTimestamp(duration float64) float64 {
	if duration < 10 {
		return duration / 2
	}
	return duration * 0.30
}

// convertOne turns one video into an .m4a with embedded cover art.
func convertOne(video string) (string, error) {
	output := strings.TrimSuffix(video, filepath.Ext(video)) + ".m4a"
	if _, err := os.Stat(output); err == nil {
		return output, fmt.Errorf("%s already exists, skipping", output)
	}

	dur, err := videoDuration(video)
	if err != nil {
		return output, err
	}
	codec, err := audioCodec(video)
	if err != nil {
		return output, err
	}

	// The cover goes in a temp folder so we never leave stray .jpg files behind.
	tmp, err := os.MkdirTemp("", "vidtune-")
	if err != nil {
		return output, err
	}
	defer os.RemoveAll(tmp)
	cover := filepath.Join(tmp, "cover.jpg")

	ts := strconv.FormatFloat(pickTimestamp(dur), 'f', 2, 64)

	// Grab one frame, scaled down so it is at most 1000px wide.
	if err := runFFmpeg(
		"-ss", ts, "-i", video,
		"-frames:v", "1",
		"-vf", "scale='min(1000,iw)':-2",
		"-q:v", "2",
		cover,
	); err != nil {
		return output, fmt.Errorf("extracting cover: %w", err)
	}

	// AAC can be copied as-is (fast, lossless). Anything else gets re-encoded.
	audioArgs := []string{"-c:a", "copy"}
	if codec != "aac" {
		audioArgs = []string{"-c:a", "aac", "-b:a", "192k"}
	}

	args := []string{"-i", video, "-i", cover, "-map", "0:a:0", "-map", "1:v"}
	args = append(args, audioArgs...)
	args = append(args, "-c:v", "copy", "-disposition:v:0", "attached_pic", output)

	if err := runFFmpeg(args...); err != nil {
		os.Remove(output) // don't leave a half-written file
		return output, fmt.Errorf("building audio: %w", err)
	}
	return output, nil
}