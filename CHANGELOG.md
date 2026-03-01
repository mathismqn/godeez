# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.0] - 2026-03-01

### Added

- Automatic database migration from `~/.godeez/tracks.db` to `~/Music/GoDeez/.tracks.db`.

### Changed

- Configuration now uses `DEEZER_ARL` environment variable (replaces `config.toml`).
- Database moved from `~/.godeez/tracks.db` to `~/Music/GoDeez/.tracks.db`.
- Show warning count in download summary.

### Removed

- `config.toml` configuration file and `~/.godeez` directory.
- `--config` flag from CLI.
- `secret_key` and `output_dir` configuration options.
- Watcher feature (`watch` subcommands).

### Fixed

- Track number zero-padding for correct file sorting.

## [1.3.0] - 2025-09-11

### Added

- Add new `track` command to download individual songs.
- Add `--genre` flag to fetch and embed genre information into file metadata tags.

### Fixed

- Handle empty media resources gracefully to prevent crashes.

## [1.2.0] - 2025-08-18

### Added

- Add new `artist` command to download an artist’s top tracks.
- Add `--limit` flag for the `artist` command to restrict the number of tracks.
- Add `--strict` flag for downloads: fail if the requested quality is unavailable.

### Changed

- Set default download quality to **MP3 320kbps**.

### Removed

- Remove `--quality=best` option. Fallback to lower quality is now the default behavior; use the `--strict` flag to prevent fallback.

### Fixed

- Handle error when `SNG_CONTRIBUTORS` metadata is empty.

## [1.1.1] - 2025-06-16

### Fixed

- Restore ability to download tracks without a Deezer Premium account (limited to **MP3 128kbps** for free accounts).

## [1.1.0] - 2025-05-19

### Added

- Support downloading full albums and playlists with more than 40 tracks (previous limit removed).
- Fetch and embed **BPM** and **musical key** into metadata tags.
- Add local **database system** (`tracks.db`) to track downloaded files and avoid re-downloading, even if files are renamed or moved.
- Improve CLI **output formatting** for a cleaner and more informative user experience.

### Changed

- Replace the `godeez` file in the user’s home directory with a `.godeez/` directory, which now stores both `config.toml` and `tracks.db`.  
  👉 If upgrading, move your existing config into `.godeez/config.toml`.
- Simplify `config.toml`: remove the need for `iv` and `license_token`.
- Automatically delete corrupted or incomplete files on download failure.

## [1.0.0] - 2024-10-15

### Added

- Initial release of **GoDeez** with basic Deezer album and playlist downloading capabilities.
