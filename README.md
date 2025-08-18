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
[License](#license)

</div>

## Features

* Download playlists, albums, and artists' top tracks from Deezer
* Select audio quality: MP3 128kbps, MP3 320kbps (default), or FLAC (⚠️ non-premium accounts are limited to 128kbps)
* Automatically adds metadata tags to downloaded files
* Fetch and tag songs with BPM and musical key
* Smart skip system: avoids re-downloading already existing files using hashes and metadata
* Cross-platform support (works on Windows, macOS, and Linux)
* Simple and easy-to-use CLI

## Installation

To install **GoDeez**, simply download the latest binary for your platform from the Releases page.

1. Go to the [Releases](https://github.com/mathismqn/godeez/releases) page.
2. Download the appropriate binary for your operating system (Windows, macOS, or Linux).
3. Move the binary to a directory included in $PATH for easy access (optional but recommended).

Example (Linux/macOS):
```bash
# Move the downloaded binary to /usr/local/bin for easy access from anywhere
mv godeez-1.2.0-linux-amd64 /usr/local/bin/godeez
```

## Configuration

The first time you run **GoDeez**, a configuration directory named `.godeez` will be automatically created in your home directory (`$HOME` on Linux/macOS, `%USERPROFILE%` on Windows).

Inside this directory:
- `config.toml`: main configuration file which contains several important variables that you need to fill out manually
- `tracks.db`: internal database used to track downloaded files and prevent duplicates

### Steps to configure

1. Run the application for the first time: this creates the `.godeez` directory and the `config.toml` file inside it.
2. Edit the `config.toml` file with a text editor to set the required values.

### Variables to configure

Here are the key variables you need to set in `config.toml`:

1. `arl_cookie`
* **What is it?**: The `arl_cookie` is a session cookie used for authentication with Deezer. Without this cookie, the downloader cannot access your account to retrieve playlists, albums, or songs.
* **How to retrieve it**:
	1.	Open your browser and log in to your Deezer account.
	2.	Open the Developer Tools (right-click on the page and select “Inspect” or press F12).
	3.	Navigate to the Application tab (in Chrome/Edge) or Storage tab (in Firefox).
	4.	In the left panel, look for Cookies and select `https://www.deezer.com`.
	5.	Find the arl cookie and copy its value.

2. `secret_key`
* **What is it?**: The `secret_key` is a cryptographic value used to decrypt Deezer’s media files.
* **How to retrieve it?**: While we cannot provide the specific secret_key in this documentation, it can be found online through various sources or developer communities that focus on Deezer.

3. `output_dir` (optional)
* **What is it?**: The `output_dir` is the path where downloaded music files will be saved.
* **Default**: If left empty, it defaults to `~/Music/GoDeez`.
* **Note**: Once set, it's recommended not to change it, as this may interfere with the skip system that relies on consistent file paths and hash indexing to detect already downloaded songs.

### Example

Here's an example of a minimal `config.toml` you can customize:
```toml
# ~/.godeez/config.toml

arl_cookie = 'your_arl_cookie_here'
secret_key = 'your_secret_key_here'
output_dir = ''  # optional
```

## Usage

### CLI Overview

When you run `godeez` without any additional commands, you’ll see a general help menu:
```bash
GoDeez is a tool to download music from Deezer

Usage:
  godeez [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  download    Download songs from Deezer
  help        Help about any command

Flags:
      --config string   config file (default ~/.godeez/config.toml)
  -h, --help            help for godeez

Use "godeez [command] --help" for more information about a command.
```
This provides an overview of the available commands and flags.

To download music, you need to use the download command. Here’s how the CLI looks when you run `godeez download`:
```bash
Download songs from Deezer

Usage:
  godeez download [command]

Available Commands:
  album       Download songs from an album
  artist      Download top songs from an artist
  playlist    Download songs from a playlist

Flags:
      --bpm                fetch BPM/key and add to file tags
      --config string      config file (default ~/.godeez/config.toml)
  -h, --help               help for download
  -q, --quality string     download quality [mp3_128, mp3_320, flac] (default "mp3_320")
      --strict             fail the song download if the quality is not available
  -t, --timeout duration   timeout for each download (e.g. 10s, 1m, 2m30s) (default 2m0s)

Use "godeez download [command] --help" for more information about a command.
```

## Contributing

Contributions help make **GoDeez** a better tool for everyone, and any help is greatly appreciated.
Whether it’s a bug fix, a new feature, or improving documentation, your input is valuable.

If you have an idea for improvement, feel free to fork the repository and submit a pull request. You can also open an issue if you spot a bug or have a feature suggestion.
Every bit of support counts, so don’t forget to give the project a star if you enjoy using it. Thank you for helping make this project better!

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/mathismqn/godeez/blob/main/LICENSE) file for details.

---

> ⚠️ This tool is provided for educational and personal use only. Please ensure your usage complies with Deezer’s Terms of Service.