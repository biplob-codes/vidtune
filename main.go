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

func pickTimestamp(duration float64) float64 {
	if duration < 10 {
		return duration / 2
	}
	return duration * 0.30
}

// NEW: find out which codec the first audio stream uses.
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

// NEW: run ffmpeg and include its error output if it fails.
func runFFmpeg(args ...string) error {
	cmd := exec.Command("ffmpeg", append([]string{"-y", "-loglevel", "error"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// NEW: convert one video into an .m4a with embedded cover art.
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

	fmt.Printf("Found %d video(s)\n", len(videos))

	done, failed := 0, 0
	for i, v := range videos {
		fmt.Printf("[%d/%d] %s\n", i+1, len(videos), v)
		out, err := convertOne(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, "  failed:", err)
			failed++
			continue
		}
		fmt.Println("  ->", out)
		done++
	}

	fmt.Printf("\nDone: %d converted, %d failed/skipped\n", done, failed)
	if failed > 0 {
		return fmt.Errorf("%d video(s) did not convert", failed)
	}
	return nil
}