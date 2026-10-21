package main

import (
	"fmt"
	"os"
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