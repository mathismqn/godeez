<div align="center">
  
# GoDeez

[![GitHub release](https://img.shields.io/github/v/release/mathismqn/godeez)](https://github.com/mathismqn/godeez/releases)
[![License](https://img.shields.io/github/license/mathismqn/godeez)](https://github.com/mathismqn/godeez/blob/main/LICENSE)
[![Last commit](https://img.shields.io/github/last-commit/mathismqn/godeez)](https://github.com/mathismqn/godeez/commits/main)

A simple Go tool for downloading music from [Deezer](https://www.deezer.com).

[Features](#features) •
[Installation](#installation) •
[Configuration](#configuration) •
[Usage](#usage) •
[Contributing](#contributing) •
[Support](#⭐-support-the-project) •
[License](#license)

</div>

## Features

- Download playlists, albums, artists’ top tracks, and individual tracks
- Choose audio quality: **MP3 128kbps**, **MP3 320kbps** (default), or **FLAC** (⚠️ non‑premium accounts are limited to 128kbps)
- Automatically embed metadata tags (artist, album, title, artwork, etc.)
- Fetch and tag songs with **BPM**, **musical key**, and **genre**
- Skip already-downloaded files using hashes and metadata
- Support Windows, macOS, and Linux
- Provide a simple, easy-to-use CLI

## Installation

To install **GoDeez**, download the latest binary for your platform from the [Releases](https://github.com/mathismqn/godeez/releases) page.

1. Go to the [Releases](https://github.com/mathismqn/godeez/releases) page.
2. Download the appropriate binary for your operating system (Windows, macOS, or Linux).
3. (Optional) Move the binary to a directory included in `$PATH` for easier access.

Example (Linux/macOS):

```bash
# Move the downloaded binary to /usr/local/bin for easy access from anywhere
mv godeez-1.4.0-linux-amd64 /usr/local/bin/godeez
```

## Configuration

**GoDeez** requires a Deezer ARL cookie for authentication. Set it as an environment variable:

```bash
export DEEZER_ARL=”your_arl_cookie_here”
```

To make it persistent, add the line above to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.).

### How to retrieve your ARL cookie

1. Open your browser and log in to your [Deezer](https://www.deezer.com) account.
2. Open the Developer Tools (right-click on the page and select “Inspect” or press F12).
3. Navigate to the **Application** tab (Chrome/Edge) or **Storage** tab (Firefox).
4. In the left panel, look for **Cookies** and select `https://www.deezer.com`.
5. Find the `arl` cookie and copy its value.

> **Note:** The ARL cookie may expire after some time. If you get authentication errors, retrieve a fresh cookie using the steps above.

### Output directory

Downloaded files are saved to `~/Music/GoDeez`. The download database (`.tracks.db`) is stored alongside your music in the output directory.

> **Upgrading from v1.3.0?** The `~/.godeez` directory and `config.toml` are no longer used. Set the `DEEZER_ARL` environment variable instead. Your existing database will be migrated automatically on first run.

## Usage

### CLI Overview

Running `godeez` without arguments shows the help menu:

```bash
GoDeez is a tool to download music from Deezer

Usage:
  godeez [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  download    Download songs from Deezer
  help        Help about any command

Flags:
  -h, --help   help for godeez

Use "godeez [command] --help" for more information about a command.
```

### Download commands

```bash
Download songs from Deezer

Usage:
  godeez download [command]

Available Commands:
  album       Download songs from an album
  artist      Download top songs from an artist
  playlist    Download songs from a playlist
  track       Download a single track

Flags:
      --bpm                fetch BPM/key and add to file tags
      --config string      config file (default ~/.godeez/config.toml)
      --genre              fetch genre and add to file tags
  -h, --help               help for download
  -q, --quality string     download quality [mp3_128, mp3_320, flac] (default "mp3_320")
      --strict             fail the song download if the quality is not available
  -t, --timeout duration   timeout for each download (e.g. 10s, 1m, 2m30s) (default 2m0s)

Use "godeez download [command] --help" for more information about a command.
```

### Examples

```bash
# Download an album
godeez download album 12345678

# Download a playlist
godeez download playlist 87654321

# Download top tracks from an artist
godeez download artist 11223344 --limit 5

# Download a single track
godeez download track 98765432

# Download with specific quality, BPM and genre data
godeez download track 98765432 --quality flac --bpm --genre
```

## Contributing

Contributions help make **GoDeez** a better tool for everyone, and any help is greatly appreciated.
Whether it’s a bug fix, a new feature, or improving documentation, your input is valuable.

If you have an idea for improvement, feel free to fork the repository and submit a pull request. You can also open an issue if you spot a bug or have a feature suggestion.

## ⭐ Support the Project

If **GoDeez** helps you enjoy your music collection, please consider giving it a star!

**Why star us?**

- Helps more music lovers discover the project
- Shows appreciation for the work and motivates development
- Takes just one click but means the world to us!

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/mathismqn/godeez/blob/main/LICENSE) file for details.

---

> ⚠️ This tool is provided for educational and personal use only. Please ensure your usage complies with Deezer’s Terms of Service.
