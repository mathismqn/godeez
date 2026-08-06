<div align="center">
  
# GoDeez

[![GitHub release](https://img.shields.io/github/v/release/mathismqn/godeez)](https://github.com/mathismqn/godeez/releases)
[![License](https://img.shields.io/github/license/mathismqn/godeez)](https://github.com/mathismqn/godeez/blob/main/LICENSE)
[![Last commit](https://img.shields.io/github/last-commit/mathismqn/godeez)](https://github.com/mathismqn/godeez/commits/main)

Download music from [Deezer](https://www.deezer.com) in MP3 or lossless FLAC/WAV.

[Features](#features) •
[Installation](#installation) •
[Updating](#updating) •
[Configuration](#configuration) •
[Usage](#usage)

</div>

## Features

- Download playlists, albums, artists' top tracks, and individual tracks
- Choose audio quality: MP3 128 kbps, MP3 320 kbps (default), or lossless FLAC/WAV (⚠️ non-premium accounts are limited to MP3 128 kbps)
- Authenticate with an ARL cookie or with your Deezer email and password
- Automatically embed metadata tags (artist, album, title, artwork, etc.)
- Fetch and tag tracks with BPM, musical key, and genre
- Works on Windows, macOS, and Linux

## Installation

Prebuilt binaries are available for every release.

1. Go to the [Releases](https://github.com/mathismqn/godeez/releases) page.
2. Download the appropriate binary for your operating system and architecture, named `godeez_<version>_<os>_<arch>`.
3. (Optional) Move the binary to a directory on your `$PATH` for easier access.

Example (Linux/macOS):

```bash
# Make it executable and move it to /usr/local/bin for access from anywhere
chmod +x godeez_1.5.0_linux_amd64
mv godeez_1.5.0_linux_amd64 /usr/local/bin/godeez
```

Each release also includes a `checksums.txt`, so you can verify your download:

```bash
sha256sum -c checksums.txt --ignore-missing
```

### macOS

The macOS binaries are not signed with an Apple Developer certificate, so
Gatekeeper blocks them on first run. You only need to clear the quarantine flag
once:

```bash
xattr -d com.apple.quarantine /usr/local/bin/godeez
```

## Updating

**GoDeez** can update itself to the latest release:

```bash
# Check for a new version
godeez update --check

# Download, verify, and install it
godeez update

# Reinstall even if already up to date
godeez update --force
```

The new binary is verified against the release's published SHA256 checksum
before it replaces the current one. If **GoDeez** lives in a directory you do
not own (such as `/usr/local/bin` on some systems), run `sudo godeez update`.

To disable new-version notifications:

```bash
export GODEEZ_NO_UPDATE_CHECK=1
```

To check which version you are running:

```bash
godeez version
```

## Configuration

**GoDeez** authenticates to Deezer in one of two ways: with an **ARL cookie** copied from your browser, or with your **email and password**. The ARL cookie works out of the box and is the recommended option; email/password login requires two extra keys that **GoDeez** does not ship (see below).

### ARL cookie

Set your ARL cookie as an environment variable:

```bash
export DEEZER_ARL="your_arl_cookie_here"
```

To make it persistent, add the line above to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.).

#### How to retrieve your ARL cookie

1. Open your browser and log in to your [Deezer](https://www.deezer.com) account.
2. Open the **Developer Tools** (right-click on the page and select **Inspect**, or press <kbd>F12</kbd>).
3. Navigate to the **Application** tab (Chrome/Edge) or **Storage** tab (Firefox).
4. In the left panel, look for **Cookies** and select **https://www.deezer.com**.
5. Find the `arl` cookie and copy its value.

> **Note:** The ARL cookie may expire after some time. If you get authentication errors, retrieve a fresh cookie using the steps above.

### Email and password

Instead of copying a cookie, you can log in once with your Deezer account:

```bash
godeez login
```

You will be prompted for your email and password. On success, **GoDeez** stores your credentials in your system keyring under the service name `godeez`. From then on, **GoDeez** authenticates on its own and renews the session when it expires.

To remove the stored credentials:

```bash
godeez logout
```

#### Gateway keys

Email/password login goes through Deezer's mobile gateway, which requires two keys:

```bash
export DEEZER_MOBILE_API_KEY="your_api_key_here"
export DEEZER_MOBILE_GW_KEY="your_gateway_key"   # exactly 16 characters
```

**GoDeez** does not bundle these keys, so you have to supply your own. For background on what they are and where they live in Deezer's clients, see [this write-up](https://gist.github.com/svbnet/b79b705a4c19d74896670c1ac7ad627e). If either variable is missing, `godeez login` exits with an error.

> **Note:** `DEEZER_ARL` takes precedence over stored credentials. If it is set, **GoDeez** always uses the cookie and never falls back to your login, so unset it (and remove it from your shell profile) before running `godeez login`.

> **Note:** The keyring entry holds your password alongside the ARL because the password is reused to renew expired sessions. On Linux, the keyring requires a running secret service; without one, `godeez login` fails with `system keyring is unavailable`.

### Output directory

Downloaded files are saved to `~/Music/GoDeez`. The download database (`.tracks.db`) is stored in the same directory as your music.

> **Upgrading from v1.3.0?** The `~/.godeez` directory and `config.toml` are no longer used. Set the `DEEZER_ARL` environment variable instead. Your existing database will be migrated automatically on first run.

## Usage

### CLI overview

Running `godeez` without arguments shows the help menu:

```text
GoDeez is a tool to download music from Deezer

Usage:
  godeez [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  download    Download tracks from Deezer
  help        Help about any command
  login       Log in to Deezer with your email and password
  logout      Remove stored Deezer credentials
  update      Update GoDeez to the latest version
  version     Print the current version of GoDeez

Flags:
  -h, --help   help for godeez

Use "godeez [command] --help" for more information about a command.
```

### Download commands

```text
Download tracks from Deezer

Usage:
  godeez download [command]

Available Commands:
  album       Download tracks from an album
  artist      Download an artist's top tracks
  playlist    Download tracks from a playlist
  track       Download a single track

Flags:
      --bpm                fetch BPM/key and add to file tags
      --genre              fetch genre and add to file tags
  -h, --help               help for download
  -q, --quality string     download quality [mp3_128, mp3_320, flac, wav] (default "mp3_320")
      --strict             fail the download if the requested quality is unavailable
  -t, --timeout duration   timeout for each download (e.g. 10s, 1m, 2m30s) (default 2m0s)

Use "godeez download [command] --help" for more information about a command.
```

> **Note:** The `artist` command takes an extra `-l, --limit` flag to choose how many top tracks to download (default 10, maximum 100).

### Examples

```bash
# Download an album
godeez download album 12345678

# Download a playlist
godeez download playlist 87654321

# Download an artist's top tracks (limit to 5 tracks)
godeez download artist 11223344 --limit 5

# Download a single track
godeez download track 98765432

# Download with specific quality, BPM, and genre data
godeez download track 98765432 --quality flac --bpm --genre
```

## Contributing

Contributions make **GoDeez** better for everyone, and any help is greatly appreciated — whether it's a bug fix, a new feature, or a documentation improvement.

To contribute, fork the repository and open a pull request. To report a bug or suggest a feature, open an issue instead.

## Support the project

If **GoDeez** helps you enjoy your music collection, please consider giving it a star ⭐

**Why star the project?**

- Helps more music lovers discover it
- Shows appreciation for the work and keeps me motivated
- Takes one click, and it means a lot

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/mathismqn/godeez/blob/main/LICENSE) file for details.

---

> ⚠️ This tool is provided for educational and personal use only. Please ensure your usage complies with Deezer's Terms of Service.
