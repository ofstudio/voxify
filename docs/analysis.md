# Product and System Analysis

This document describes the current Voxify implementation for developers maintaining the project. It covers the working tree, including the responsive card layout and documented Makefile, rather than asserting what is deployed on a particular server. It is an analysis of existing behavior, not a roadmap or a specification for new features.

Read [Architecture and Operations](architecture.md) for component contracts, execution flow, configuration, and deployment. The [README](../README.md) provides the installation overview. Source and test links below identify the evidence for behavior; a test name alone does not establish a stronger guarantee than its assertions.

## Purpose and system boundary

Voxify turns supported YouTube video links into downloadable audio episodes and publishes a podcast RSS feed and a static landing page. Its intended setting is a personal installation or a small group sharing one collection.

| Actor | Responsibility and access |
| --- | --- |
| Operator | Configures the bot, allowed user IDs, feed metadata, storage, external tools, and public hosting. |
| Allowed Telegram user | Submits links and requests feed information or regeneration. All allowed users contribute to the same feed. |
| Listener | Reads the landing page or subscribes through a podcast player. No Telegram identity is required to access generated files. |
| External systems | Telegram transports commands and replies; YouTube provides source content; yt-dlp and ffmpeg retrieve and process it. |

The application maintains one feed per deployment. There are no user-owned feeds, episode ownership rules, web administration API, or embedded HTTP file server. Telegram authorization applies to bot interaction; public file access is controlled by the hosting layer. See [application assembly](../internal/app/app.go), [Telegram handlers](../internal/handlers/telegram.go), and the [Compose example](../docker-compose-example.yaml).

## User scenarios

| Scenario | Preconditions and trigger | Current result | Failure or alternative |
| --- | --- | --- | --- |
| Get introduction | Allowed user sends `/start`. | Bot sends usage instructions. | Unauthorized updates are blocked without a user-facing denial message. |
| Add an episode | Allowed user sends message text beginning with `https://`; URL and configured format/quality pass validation; a worker is ready. | Bot reports download start, retrieves content, stores the episode, and reports download success; feed regeneration is then requested. | Invalid/unsupported input, duplicates, busy workers, tool errors, and interruption produce mapped error replies. |
| Inspect feed | Allowed user sends `/info`. | Bot returns configured feed metadata, episode count, RSS URL, and retention limit when applicable. | A store or template error produces an error reply. This command does not check whether the public HTTP endpoint is reachable. |
| Rebuild publications | Allowed user sends `/build`. | Landing page is regenerated; RSS is regenerated if episodes exist. | Build errors are reported. An empty database still produces a successful build response, although RSS generation is skipped. |
| Browse episodes | Listener opens the public landing page. | All stored episodes appear as cards; cards link to the original content source. | Availability depends on hosting and existing public files. The landing page is not an embedded audio player. |
| Subscribe and listen | Listener adds the public RSS URL to a podcast player. | The player fetches RSS and downloads enclosure URLs served from the public directory. | Removed episodes have no application-provided redirect or archive. Player refresh timing is outside Voxify. |
| Restart the application | Operator starts the configured service. | Dependencies are checked, temporary downloads are cleared, retention is applied, and publications are rebuilt before bot startup. | Initialization errors stop startup. In-flight requests are not restored. |

Sources: [command handling](../internal/handlers/telegram.go), [request handling](../internal/handlers/request.go), [notifications](../internal/handlers/notification.go), [publication service](../internal/services/feed.go), [service initialization](../internal/services/container.go).

### Input and conversion rules

- The Telegram URL handler passes the entire message text as the URL. It does not extract multiple links, trim surrounding prose, or implement batch submissions.
- The installed platform matches these exact, case-sensitive prefixes: `https://www.youtube.com/`, `https://youtube.com/`, `https://m.youtube.com/`, and `https://youtu.be/`. yt-dlp's broader platform support is not exposed automatically.
- Service validation accepts an `http` or `https` scheme, but Telegram routing and the installed platform restrict the usable path to the HTTPS prefixes above.
- Only `mp3` and `m4a` pass format validation. Quality must match `^[0-9a-zA-Z-_]{1,32}$`; semantic interpretation is delegated to yt-dlp.
- Conversion uses `--no-playlist`: there is no playlist import workflow. Metadata and audio are fetched separately. An available thumbnail is cropped to a square and resized through ffmpeg; failure to process that thumbnail fails the download.
- If source metadata has no title, the uploader is used. An empty description becomes `-`. Audio extraction requests embedded thumbnails and metadata from yt-dlp.

Sources: [validation and platform selection](../internal/services/episode.go), [yt-dlp adapter](../internal/platforms/yt-dlp.go). Validation is exercised in [episode service tests](../internal/services/episode_test.go) using platform mocks.

### Admission, duplicates, and status

`DOWNLOAD_WORKERS` controls concurrent downloads. The request channel is unbuffered and admission uses a nonblocking send. A request is accepted only when a worker can receive it immediately; otherwise it fails with `ErrDownloadBusy`. There is no backlog to wait in and no automatic application-level retry.

An in-memory map protects active downloads by the exact submitted URL. Stored duplicates are checked by equality against either `original_url` or `canonical_url`, before dispatch and again during download. Different URL aliases for the same video are not normalized before admission, and the database has no unique constraint on these columns. These checks therefore do not guarantee one record per underlying video under concurrent alias submissions.

`pending`, `success`, and `failed` are response-event statuses, not durable job states. Download success means that media retrieval, database insertion, and retention enforcement returned successfully. It does **not** mean that the subsequent RSS build succeeded. Automatic builds have no request source, so their results are logged without a Telegram build reply. The asynchronous event bus also does not guarantee notification delivery order.

Sources: [request handlers](../internal/handlers/request.go), [response types](../internal/domain/responses.go), [notifications](../internal/handlers/notification.go), [schema](../internal/store/migration/01_initial.up.sql). Related tests: [request handlers](../internal/handlers/request_test.go), [notifications](../internal/handlers/notification_test.go).

## Data model

### Persisted episode

There is one application table, `episodes`. It contains metadata and relative filenames; the audio and artwork bytes live in the filesystem. There is no persisted user, feed configuration, job, or request table.

| Domain field / database column | Meaning |
| --- | --- |
| `ID` / `id` | SQLite-generated integer primary key; separate from the transient request ID used in filenames. |
| `Title` / `title` | Source title, with uploader fallback. |
| `Description` / `description` | Source description; the full value is retained independently of landing-page truncation. |
| `ThumbnailFile` / `thumbnail_file` | Relative JPEG filename when source artwork is available. |
| `MediaFile` / `media_file` | Relative audio filename, based on request ID and output format. |
| `MediaType` / `media_type` | Enclosure MIME type: `audio/mpeg` or `audio/x-m4a`. |
| `MediaDuration` / `media_duration` | Duration from source metadata, in seconds. |
| `MediaSize` / `media_size` | Output audio file size, in bytes. |
| `Author` / `author` | Source uploader, distinct from configured feed author. |
| `OriginalURL` / `original_url` | Exact URL submitted for the download. |
| `CanonicalURL` / `canonical_url` | `webpage_url` returned by yt-dlp. |
| `CreatedAt` / `created_at` | SQLite insertion timestamp; this becomes the episode publication date. It is not the source video's original publication date. |

The schema has non-unique indexes on original URL, canonical URL, and creation timestamp. Episodes are read in descending `created_at` order and removed in ascending order; there is no secondary ordering key for equal timestamps. Optional SQL columns and Go zero values should be interpreted from the schema and store implementation rather than inferred from this business-level table.

Sources: [Episode](../internal/domain/episode.go), [migration](../internal/store/migration/01_initial.up.sql), [SQLite store](../internal/store/sqlite.go), [store tests](../internal/store/sqlite_test.go).

### Derived feed and publications

`FeedInfo` combines process configuration with a database count and the latest episode timestamp. It is not stored separately. Feed title, description, author, language, categories, artwork, keywords, and explicit flag come from configuration. Each category list uses its first element as the parent and the remaining elements as subcategories. Owner, copyright, feed type, completed, and blocked capabilities exist in types but are not configured by the application service.

RSS enclosure URLs and GUIDs use `PUBLIC_URL` joined with the media filename. Changing the public base URL consequently changes generated episode GUIDs. RSS contains all retained episodes. The landing page also renders all retained episodes, with descriptions shortened to at most 128 Unicode code points including the ellipsis. Card links prefer canonical URL, then original URL, then `#`. Cards use shrinkable grid tracks and emergency text wrapping for long URLs; metadata can wrap onto another line.

Sources: [FeedInfo](../internal/domain/feed-info.go), [feed generation](../internal/services/feed.go), [landing template](../internal/templates/landing.go), [template helpers](../internal/templates/func.go). Related tests: [feed service](../internal/services/feed_test.go), [template helpers](../internal/templates/func_test.go).

## Retention and deletion

`FEED_MAX_EPISODES > 0` limits stored episodes. Zero is the documented unlimited setting; the implementation also treats negative values as unlimited. Enforcement runs during episode-service initialization and after each successful database insertion, not as a separate scheduled task or as part of `/build`.

For each excess oldest episode, the service deletes audio and thumbnail files before deleting the database record. Missing files are tolerated. Other deletion errors stop enforcement. Absolute paths and paths escaping through `..` are rejected by the deletion helper. There is no recycle bin or retained media archive.

A retention error after insertion makes the download request fail even though the new episode may already exist. File and database deletion are not a single transaction; partial deletion can leave records referring to missing files. These are consequences of the current operation order, not guarantees of recoverable rollback.

Sources: [retention implementation](../internal/services/episode.go), [retention tests](../internal/services/episode_test.go).

## Error interpretation and maintenance implications

| Condition | Observable behavior | Maintenance implication |
| --- | --- | --- |
| Unauthorized Telegram user | Update is logged and blocked. | Verify allowed IDs rather than expecting a denial reply. |
| Busy worker pool | Immediate busy reply. | Retry the submission later; no job has been queued for future execution. |
| Exact URL active or stored | In-progress or already-downloaded reply. | An existing episode can outlive a failed publication; inspect the feed separately. |
| Download timeout or external command error | Download-failed reply; command detail in logs. | Check source availability, tool versions, and configured timeout. A timeout alone is not mapped to the interruption message. |
| Application context canceled during download | Failure is wrapped as interrupted. | The notification itself may fail because it uses the canceled context. |
| Database insertion fails after file publication | Service attempts to delete the downloaded files. | Cleanup failures are logged; orphaned files are possible. |
| Automatic publication fails | Build failure is logged; no automatic build reply. | Fix the underlying problem and use `/build` to regenerate from stored episodes. |
| Build with zero episodes | Empty landing page replaces the prior page; RSS creation is skipped. | An existing RSS file is not removed and can remain stale. |

Sources: [error definitions](../internal/domain/errors.go), [message mapping](../internal/templates/messages.go), [episode service](../internal/services/episode.go), [feed service](../internal/services/feed.go), [notification handlers](../internal/handlers/notification.go).

## Current operating limits

The design uses one process, local storage, and an in-memory event bus. It has no persistent retry scheduler, cross-instance download coordination, publication-version coordination, application health endpoint, or built-in metrics exporter. There is no measured throughput or availability objective specified in the repository. External content availability and player behavior are outside the application's guarantees.

The implementation supports a small shared installation; scaling multiple processes against the same directories introduces uncoordinated cleanup and publication. See [consistency and operations](architecture.md) for the concrete boundaries and recovery procedures. These observations document the current system and do not introduce a target architecture.
