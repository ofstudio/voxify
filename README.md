![Voxify logo](assets/voxify-logo.svg)

# Voxify

A Telegram bot that converts YouTube videos and other online content into podcast feeds. Voxify allows you to create
personalized RSS podcast feeds from your favorite video content, making it easy to listen to educational videos,
interviews, and other content in podcast format.

## About

Voxify transforms your favorite video content into audio-only podcast episodes that you can listen to anywhere. Simply
send YouTube links or other supported video URLs to the Telegram bot, and it will:

- Download the video content using yt-dlp
- Extract high-quality audio in your preferred format (MP3, M4A, etc.)
- Generate episode thumbnails and metadata
- Add episodes to your personal RSS podcast feed
- Make the content accessible through any podcast player

The bot maintains a personal RSS feed for each user, allowing you to build a custom podcast library from diverse video
sources. Perfect for listening to educational content, tech talks, interviews, or any video content during commutes,
workouts, or while multitasking.

## Usage

1. **Get your Telegram user ID**: Message [@userinfobot](https://t.me/userinfobot) on Telegram to find your user ID
2. **Create a Telegram bot**: Message [@BotFather](https://t.me/BotFather) to create a new bot and get the bot token
3. **Configure and deploy** the Voxify bot (see Installation section below)
4. **Start using the bot**:
    - Send YouTube or other supported video URLs to your bot
    - The bot will process the video and add it to your podcast feed
    - Access your RSS feed at `https://yourdomain.com/rss.xml`
    - Subscribe to the feed in your favorite podcast player

### Currently Supported Platforms

- YouTube

More platforms will be added in future releases.

## Installation

### Using Docker Compose (Recommended)

The easiest way to deploy Voxify is using the pre-built Docker image with Docker Compose.

1. **Create a data directory**:
   ```bash
   mkdir -p /path/to/voxify-data/{db,public,downloads}
   ```

2. **Download the example docker-compose.yaml**:
   ```bash
   wget https://raw.githubusercontent.com/ofstudio/voxify/main/docker-compose-example.yaml -O docker-compose.yaml
   ```

3. **Create environment file**:
   ```bash
   wget https://raw.githubusercontent.com/ofstudio/voxify/main/.env-example -O .env
   ```

4. **Edit the configuration**:
    - Update `.env` with your bot token, user IDs, and other settings
    - Modify `docker-compose.yaml` to replace placeholders with your actual values:
        - Replace `</path/to/data>` with your actual data directory path
        - Replace `<mydomain.com>` with your domain name
        - Replace `<my@email.org>` with your email for Let's Encrypt
        - Set `VERSION` to the latest release version

5. **Deploy**:
   ```bash
   docker-compose up -d
   ```

The example includes Traefik reverse proxy and Nginx for serving files over HTTPS with automatic Let's Encrypt
certificates.

### Available Docker Images

Pre-built images are available at:

- `ghcr.io/ofstudio/voxify:latest` - Latest stable release
- `ghcr.io/ofstudio/voxify:v1.x.x` - Specific version tags

Check [packages](https://github.com/ofstudio/voxify/pkgs/container/voxify) for the latest version.

## Environment Variables

| Variable                 | Description                                                                                                                                                                     |
|--------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `TELEGRAM_BOT_TOKEN`     | **Required.** Telegram bot token from [@BotFather](https://t.me/BotFather). Example: `123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZ`                                                     |
| `TELEGRAM_ALLOWED_USERS` | **Required.** Comma-separated list of allowed Telegram user IDs. Example: `123456789,987654321`. Note: you can find your user ID using [@userinfobot](https://t.me/userinfobot) |
| `PUBLIC_URL`             | **Required.** Public URL where feed and media files are accessible. Example: `https://example.com/podcasts`                                                                     |
| `DB_FILEPATH`            | **Required.** Path to SQLite database file. Default: `./data/voxify.db`                                                                                                         |
| `PUBLIC_DIR`             | **Required.** Path to public directory for feed and media files. Default: `./data/public`                                                                                       |
| `DOWNLOAD_DIR`           | **Required.** Path to temporary download directory. Default: `./data/downloads`                                                                                                 |
| `DOWNLOAD_TIMEOUT`       | *Optional.* Timeout for downloading media. Default: `1h` (formats: 30s, 10m, 1h)                                                                                                |
| `DOWNLOAD_FORMAT`        | *Optional.* Media download format. Default: `mp3` (options: mp3, m4a, etc.)                                                                                                     |
| `DOWNLOAD_QUALITY`       | *Optional.* Audio quality for downloaded media. Default: `192k`                                                                                                                 |
|  `DOWNLOAD_WORKERS`      | *Optional.* Number of concurrent download workers. Default: `2`                                                                                                                 |
| `THUMBNAIL_SIZE`         | *Optional.* Size of square thumbnail in pixels. Default: `3000`                                                                                                                 |
| `YT_DLP_PATH`            | *Optional.* Path to yt-dlp executable. Default: `yt-dlp`                                                                                                                        |
| `FFMPEG_PATH`            | *Optional.* Path to ffmpeg executable. Default: `ffmpeg`                                                                                                                        |
| `FEED_FILENAME`          | *Optional.* Name of the RSS feed file. Default: `rss.xml`                                                                                                                       |
| `FEED_TITLE`             | *Optional.* Title of the RSS feed. Default: `Voxify Podcast`                                                                                                                    |
| `FEED_DESC`              | *Optional.* Description of the RSS feed. Default: `Voxify Podcast description`                                                                                                  |
| `FEED_IMAGE`             | *Optional.* URL of the RSS feed cover image. Example: `https://example.com/cover.jpg`                                                                                           |
| `FEED_LANGUAGE`          | *Optional.* Language code for the RSS feed. Default: `en`                                                                                                                       |
| `FEED_CATEGORIES`        | *Optional.* Primary categories (comma-separated). Default: `Technology`                                                                                                         |
| `FEED_CATEGORIES2`       | *Optional.* Additional categories (comma-separated). Example: `Science,Astronomy`                                                                                               |
| `FEED_CATEGORIES3`       | *Optional.* Additional categories (comma-separated). Example: `Education`                                                                                                       |
| `FEED_IS_EXPLICIT`       | *Optional.* Whether feed contains explicit content. Default: `false` (options: true, false)                                                                                     |
| `FEED_AUTHOR`            | *Optional.* Author of the RSS feed. Example: `John Doe`                                                                                                                         |
| `FEED_LINK`              | *Optional.* Link to the website of the RSS feed. Default: `https://github.com/ofstudio/voxify`                                                                                  |
| `FEED_KEYWORDS`          | *Optional.* Comma-separated keywords for the RSS feed. Example: `podcast,tech,news,interviews`                                                                                  |

## Acknowledgments

- Built with [Go](https://golang.org/)
- Uses [yt-dlp](https://github.com/yt-dlp/yt-dlp) for video downloading
- Telegram bot powered by [go-telegram/bot](https://github.com/go-telegram/bot)

## License

Apache License 2.0

## Contributing

Feel free to open an issue or a pull request.

## Author

Oleg Fomin [@ofstudio](https://t.me/ofstudio)
