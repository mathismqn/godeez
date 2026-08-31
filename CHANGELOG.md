# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.1] - 2026-08-31

### Changed

- Match candidate files by size before hashing, so a moved or renamed download is found without re-hashing the library.
- Report progress while the download database is updated on the first run after upgrading.

### Fixed

- Download tracks that failed with `invalid track token` by following the entry Deezer plays them from ([#8](https://github.com/mathismqn/godeez/issues/8)).
- Use the artwork of the linked entry when a track only has a grey placeholder cover.
- Prefix and tag album tracks with their disc number so multi-disc albums no longer repeat the first track number ([#9](https://github.com/mathismqn/godeez/issues/9)).
- Skip personal uploads in playlists instead of failing the whole playlist, as Deezer serves them with no streaming rights.

## [1.5.0] - 2026-08-05

### Added

- Add new `login` and `logout` commands to authenticate with your Deezer email and password. Credentials are stored in the system keyring. Requires `DEEZER_MOBILE_API_KEY` and `DEEZER_MOBILE_GW_KEY` to be set.
- Add new `update` command to replace the binary in place with the latest release, with `--check` to only report availability and `--force` to reinstall.
- Add new `version` command to print the version, commit, build date, and platform.
- Notify when a newer version is available after a download completes.
- Add WAV download quality (`--quality=wav`): the FLAC stream is converted locally to lossless WAV.

### Changed

- `DEEZER_ARL` is now optional. When it is unset, the credentials stored by `godeez login` are used instead, and expired sessions are renewed automatically.
- Release binaries are now named `godeez_<version>_<os>_<arch>` (previously `godeez-<version>-<os>-<arch>`) and are published alongside a `checksums.txt` file.

### Fixed

- Interrupted downloads no longer leave partial files behind.
- Avoid overwriting an existing file when another track resolves to the same name.
- Report a clear error when the database is already in use by another process.
- Migrate the legacy database when `~/.godeez` and `~/Music/GoDeez` are on different filesystems.
- Write metadata tags even when the cover art or track duration is missing.

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

- Add new `track` command to download individual tracks.
- Add `--genre` flag to fetch and embed genre information into file metadata tags.

### Fixed

- Handle empty media resources gracefully to prevent crashes.

## [1.2.0] - 2025-08-18

### Added

- Add new `artist` command to download an artist’s top tracks.
- Add `--limit` flag for the `artist` command to restrict the number of tracks.
- Add `--strict` flag for downloads: fail if the requested quality is unavailable.

### Changed

- Set default download quality to MP3 320 kbps.

### Removed

- Remove `--quality=best` option. Fallback to lower quality is now the default behavior; use the `--strict` flag to prevent fallback.

### Fixed

- Handle error when `SNG_CONTRIBUTORS` metadata is empty.

## [1.1.1] - 2025-06-16

### Fixed

- Restore ability to download tracks without a Deezer Premium account (limited to MP3 128 kbps for free accounts).

## [1.1.0] - 2025-05-19

### Added

- Support downloading full albums and playlists with more than 40 tracks (previous limit removed).
- Fetch and embed BPM and musical key into metadata tags.
- Add local database system (`tracks.db`) to track downloaded files and avoid re-downloading, even if files are renamed or moved.
- Improve CLI output formatting for a cleaner and more informative user experience.

### Changed

- Replace the `godeez` file in the user’s home directory with a `.godeez/` directory, which now stores both `config.toml` and `tracks.db`.  
  👉 If upgrading, move your existing config into `.godeez/config.toml`.
- Simplify `config.toml`: remove the need for `iv` and `license_token`.
- Automatically delete corrupted or incomplete files on download failure.

## [1.0.0] - 2024-10-15

### Added

- Initial release of **GoDeez** with basic Deezer album and playlist downloading capabilities.

[1.5.1]: https://github.com/mathismqn/godeez/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/mathismqn/godeez/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/mathismqn/godeez/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/mathismqn/godeez/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/mathismqn/godeez/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/mathismqn/godeez/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/mathismqn/godeez/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/mathismqn/godeez/releases/tag/v1.0.0
