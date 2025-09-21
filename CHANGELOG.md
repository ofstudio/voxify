# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/ofstudio/voxify/compare/v0.1.0...HEAD

[v0.1.0]: https://github.com/ofstudio/voxify/compare/v0.0.1...v0.1.0

[v0.0.1]: https://github.com/ofstudio/voxify/releases/tag/v0.0.1

[#6]: https://github.com/ofstudio/voxify/issues/6

[#5]: https://github.com/ofstudio/voxify/issues/5

[#4]: https://github.com/ofstudio/voxify/issues/4