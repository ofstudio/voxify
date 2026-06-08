# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v1.1.0] - 2026-06-08

### Added

- Optional `FEED_MAX_EPISODES` setting to keep only the newest episodes in the feed.

### Changed

- Updated `yt-dlp` to version `2026.03.17`.
- When the feed episode limit is exceeded, Voxify permanently deletes the oldest episode records and associated media/thumbnail files on startup and after new downloads.
- Documented that generated RSS, landing page and media files are public by design.
- RSS feed and landing page files are now replaced atomically.

## [v1.0.3] - 2026-02-18

### Changed

- Updated `yt-dlp` to version `2026.02.04`.

## [v1.0.2] - 2025-10-24

### Changed

- Updated `yt-dlp` to version `2025.10.22`.

## [v1.0.1] - 2025-10-04

### Fixed

- Thread-safety issue with global `rand.Source` in token generation. Resolves [#7].
- HTML formatting injections in Telegram messages by properly escaping user input. Resolves [#10].

## [v1.0.0] - 2025-10-03

### Added

- Landing page with episode cards and light/dark theme support, displaying feed information and all episodes with thumbnails, titles and  descriptions.
- Automated installation of the latest yt-dlp version via `pip` during Docker image build (instead of Alpine package).

### Changed

- Significant application architecture refactoring, improving code organization, maintainability, and testability.
- Enhanced template system with separate templates for landing page and feed information display.
- Improved error handling and logging throughout the application for better debugging and monitoring.

## [v0.1.0] - 2025-09-22

### Added

- Telegram command `/info` that shows detailed feed information. Resolves [#6].

### Fixed

- Add keywords to feed metadata if provided. Resolves [#4].
- Remove unnecessary `PORT` directive from Dockerfile. Resolves one of items from [#5].
- Clarified README about single feed vs per-user. Resolves one of items from [#5].

### Security

- Application is running under nobody user now. Resolves one of items from [#5].

## [v0.0.1] - 2025-09-20

Initial public release (tagged `v0.0.1`).

### Added

- Bot core functionality: conversion of videos to audio files and podcast feed generation.
- `/start` and `/build` commands.
- Support for YouTube links.
- Docker image and Docker Compose setup.
- Basic configuration via environment variables.

---

[Unreleased]: https://github.com/ofstudio/voxify/compare/v1.1.0...HEAD

[v1.1.0]: https://github.com/ofstudio/voxify/compare/v1.0.3...v1.1.0

[v1.0.3]: https://github.com/ofstudio/voxify/compare/v1.0.2...v1.0.3

[v1.0.2]: https://github.com/ofstudio/voxify/compare/v1.0.1...v1.0.2

[v1.0.1]: https://github.com/ofstudio/voxify/compare/v1.0.0...v1.0.1

[v1.0.0]: https://github.com/ofstudio/voxify/compare/v0.1.0...v1.0.0

[v0.1.0]: https://github.com/ofstudio/voxify/compare/v0.0.1...v0.1.0

[v0.0.1]: https://github.com/ofstudio/voxify/releases/tag/v0.0.1

[#4]: https://github.com/ofstudio/voxify/issues/4

[#5]: https://github.com/ofstudio/voxify/issues/5

[#6]: https://github.com/ofstudio/voxify/issues/6

[#7]: https://github.com/ofstudio/voxify/issues/7

[#10]: https://github.com/ofstudio/voxify/issues/10
