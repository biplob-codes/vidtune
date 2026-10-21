# vidtune

Turn music videos into audio files with cover art, straight from your terminal.

Run `vidtune make-audio` in a folder of videos and get an `.m4a` for each one, with a screenshot from the video embedded as the album cover. Players like Musicolet, Auxio, and VLC will show that cover when you play the file.

## Features

- Converts every video in the current folder in one command
- Embeds a cover image taken from the video itself
- Picks the screenshot at about 30% of the video's length, which avoids the black frames and logos that many music videos start with
- Copies AAC audio as-is (no quality loss, very fast) and re-encodes anything else to AAC 192k
- Skips files whose `.m4a` already exists, so it is safe to run twice
- Runs ffmpeg single-threaded at the lowest CPU priority, so it won't freeze your machine
- No Go dependencies, only the standard library

## Requirements

- Go 1.21 or newer
- `ffmpeg` and `ffprobe` available in your `PATH`

On Fedora:

```
sudo dnf install ffmpeg-free
```

On Ubuntu/Debian:

```
sudo apt install ffmpeg
```

On macOS:

```
brew install ffmpeg
```

## Installation

```
go install github.com/biplob-codes/vidtune@latest
```

This puts the binary in `$(go env GOPATH)/bin` (usually `~/go/bin`). Make sure that folder is on your `PATH`:

```
export PATH="$PATH:$HOME/go/bin"
```

Add that line to your `~/.bashrc` or `~/.zshrc` to make it permanent.

To build from source instead:

```
git clone https://github.com/biplob-codes/vidtune.git
cd vidtune
go install .
```

## Usage

Go to a folder containing videos and run:

```
vidtune make-audio
```

Example output:

```
Found 2 video(s)
[1/2] song-one.mp4
  -> song-one.m4a
[2/2] song-two.webm
  -> song-two.m4a

Done: 2 converted, 0 failed/skipped
```

Other commands:

```
vidtune help
```

## Supported input formats

`.mp4`, `.mkv`, `.webm`, `.mov`, `.avi`, `.m4v`

Only the current folder is scanned, not subfolders.

## How it works

For each video, vidtune:

1. Asks `ffprobe` for the video's duration and audio codec
2. Grabs one frame at 30% of the duration, scaled down to at most 1000px wide, into a temporary folder
3. Builds an `.m4a` with `ffmpeg`, taking the first audio track and attaching the frame as cover art
4. Deletes the temporary frame

Videos shorter than 10 seconds use their midpoint for the screenshot.

## Notes

- Output files are named after the video, with the `.m4a` extension, and are written next to the original.
- If a conversion fails, the partial output file is removed and vidtune continues with the next video. The exit code is non-zero if anything failed or was skipped.
- The 30% rule is a heuristic. If a particular video gets a poor cover, you can extract your own frame with ffmpeg and attach it manually.

## Project layout

```
main.go      entry point and the make-audio command
videos.go    finding video files
ffmpeg.go    ffmpeg and ffprobe helpers
convert.go   converting a single video
```

## License

MIT
