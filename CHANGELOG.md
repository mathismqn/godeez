# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2025-08-18

### Added
- New `artist` command to download an artist’s top tracks.
- `--limit` flag for the `artist` command to restrict the number of tracks.
- `--strict` flag for downloads: fail if the requested quality is unavailable.

### Changed
- Default download quality is now **MP3 320kbps**.

### Removed
- The `--quality=best` option. Fallback to lower quality is now the **default behavior**; use the new `--strict` flag if you want to prevent fallback.

### Fixed
- Handle error when `SNG_CONTRIBUTORS` metadata is empty.

## [1.1.1] - 2025-06-16

### Fixed
- Restored ability to download tracks **without a Deezer Premium account** (limited to **MP3 128kbps** for free accounts).

## [1.1.0] - 2025-05-19

### Added
- Support for downloading full albums and playlists with more than 40 tracks (previous limit removed).
- Option to fetch and embed **BPM** and **musical key** into metadata tags.
- New local **database system** (`tracks.db`) to track downloaded files and avoid re-downloading, even if files are renamed or moved.
- Improved CLI **output formatting** for a cleaner and more informative user experience.

### Changed
- The `.godeez` file in the user’s home directory has been replaced by a `.godeez/` directory.  
  It now stores both `config.toml` and the internal `tracks.db`.  
  If you're upgrading from an older version, move your existing config into `.godeez/config.toml`.
- Simplified `config.toml`: `iv` and `license_token` are no longer required.
- Cleanup logic: corrupted or incomplete files are now automatically deleted on download failure.

## [1.0.0] - 2024-10-15

### Added

- Initial release of **GoDeez** with basic Deezer album and playlist downloading capabilities.