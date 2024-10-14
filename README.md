<div align="center">
  
# GoDeez

A simple Go tool for downloading music from [Deezer](https://www.deezer.com).

[Features](#features) •
[Installation](#installation) •
[Configuration](#configuration) •
[Usage](#usage) •
[Contributing](#contributing) •
[License](#license)

</div>

## Features

* Download playlists and albums from Deezer
* Select audio quality: MP3 128kbps, MP3 320kbps, or FLAC
* Automatically adds metadata tags to downloaded files
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
mv godeez-1.0.0-linux-amd64 /usr/local/bin/godeez
```

## Configuration

The first time you run **GoDeez**, a configuration file named .godeez will be automatically generated in your home directory ($HOME on Linux/macOS, %USERPROFILE% on Windows).

This configuration file contains several important variables that you need to fill out manually. Below are the steps for retrieving and setting each variable.

### Steps to configure

1. Run the application for the first time: This generates the .godeez configuration file in your home directory.
2. Edit the configuration file: Open the .godeez file with a text editor to set the required variables.

### Variables to configure

Here are the key variables you need to set in the .godeez file:

1. `arl_cookie`
* **What is it?**: The arl_cookie is a session cookie used for authentication with Deezer. Without this cookie, the downloader cannot access your account to retrieve playlists, albums, or songs.
* **How to retrieve it**:
	1.	Open your browser and log in to your Deezer account.
	2.	Open the Developer Tools (right-click on the page and select “Inspect” or press F12).
	3.	Navigate to the Application tab (in Chrome/Edge) or Storage tab (in Firefox).
	4.	In the left panel, look for Cookies and select `https://www.deezer.com`.
	5.	Find the arl cookie and copy its value.

2. `license_token`
* **What is it?**: The license_token is required to access Deezer’s media URLs for downloading songs. This token is found in the network requests your browser makes when playing a song.
* **How to retrieve it**:
  1.	Open Developer Tools in your browser (right-click on the page and select “Inspect” or press F12).
	2.	Go to the Network tab.
	3.	Start playing a song on Deezer and look for a request to `https://media.deezer.com/v1/get_url`.
	4.	Select the request and in the Request Data section, find the license_token.
	5.	Copy the license_token value.

3. `secret_key`
* **What is it?**: The secret_key is a cryptographic value used alongside the iv to decrypt Deezer’s media files.
* **How to retrieve it?**: While we cannot provide the specific secret_key in this documentation, it can be found online through various sources or developer communities that focus on Deezer.

4. `iv`
* **What is it?**: This is another cryptographic variable needed to decrypt media streams from Deezer.
* **How to retrieve it?**: The iv can be found in the [.example-config](https://github.com/mathismqn/godeez/blob/main/.example-config) file included with this project.

## Usage

### CLI Overview

When you run **godeez** without any additional commands, you’ll see a general help menu:
```bash
GoDeez is a tool to download music from Deezer

Usage:
  godeez [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  download    Download songs from Deezer
  help        Help about any command

Flags:
      --config string   config file (default is $HOME/.godeez)
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
  album       Download songs from one or more albums
  playlist    Download songs from one or more playlists

Flags:
  -h, --help             help for download
  -o, --output string    output directory (default is current directory)
  -q, --quality string   download quality [mp3_128, mp3_320, flac, best] (default is best)

Global Flags:
      --config string   config file (default is $HOME/.godeez)

Use "godeez download [command] --help" for more information about a command.
```

## Contributing

Contributions help make **GoDeez** a better tool for everyone, and any help is greatly appreciated.
Whether it’s a bug fix, a new feature, or improving documentation, your input is valuable.

If you have an idea for improvement, feel free to fork the repository and submit a pull request. You can also open an issue if you spot a bug or have a feature suggestion.
Every bit of support counts, so don’t forget to give the project a star if you enjoy using it. Thank you for helping make this project better!

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/mathismqn/godeez/blob/main/LICENSE) file for details.
