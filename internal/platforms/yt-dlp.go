package platforms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/pkg/files"
)

// YtDlp is a services.Platform implementation using yt-dlp for downloading YouTube videos.
// It extracts media, thumbnail and metadata using yt-dlp and ffmpeg.
type YtDlp struct {
	cfg config.Settings
	log *slog.Logger
}

func NewYtDlpPlatform(cfg config.Settings, log *slog.Logger) *YtDlp {
	return &YtDlp{
		cfg: cfg,
		log: log,
	}
}

func (p YtDlp) ID() string {
	return "yt-dlp"
}

func (p YtDlp) Match(url string) bool {
	return strings.HasPrefix(url, "https://www.youtube.com/") ||
		strings.HasPrefix(url, "https://youtu.be/")
}

// Init checks if yt-dlp and ffmpeg are available and working.
func (p YtDlp) Init(_ context.Context) error {
	if p.cfg.YtDlpPath == "" {
		return fmt.Errorf("yt-dlp path is not configured")
	}
	if p.cfg.FFMpegPath == "" {
		return fmt.Errorf("ffmpeg path is not configured")
	}
	cmd := exec.Command(p.cfg.FFMpegPath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found or not working: %w", err)
	}
	cmd = exec.Command(p.cfg.YtDlpPath, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp not found or not working: %w", err)
	}
	return nil
}

const ytDlpPattern = "yt-dlp-*"

var errTempDir = errors.New("failed to create temporary directory")

func (p YtDlp) Download(ctx context.Context, req entities.Request) (*entities.Episode, error) {
	// Validate requested download format
	mediaType, supported := mediaTypes[req.DownloadFormat]
	if !supported {
		return nil, fmt.Errorf("unsupported download format: %s", req.DownloadFormat)
	}

	// Create temporary directories for downloading
	metaDir, err := os.MkdirTemp(p.cfg.DownloadDir, ytDlpPattern)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errTempDir, err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer os.RemoveAll(metaDir)
	thumbDir, err := os.MkdirTemp(p.cfg.DownloadDir, ytDlpPattern)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errTempDir, err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer os.RemoveAll(thumbDir)
	mediaDir, err := os.MkdirTemp(p.cfg.DownloadDir, ytDlpPattern)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errTempDir, err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer os.RemoveAll(mediaDir)

	// Fetch metadata
	p.log.Info("[yt-dlp] downloading metadata", "request", req.LogValue())

	meta, err := p.fetchMeta(ctx, req, metaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata from youtube: %w", err)
	}
	p.log.Info("[yt-dlp] metadata downloaded", "request", req.LogValue())

	episode := &entities.Episode{
		Title:         meta.Title,
		Description:   meta.Description,
		MediaType:     mediaType,
		MediaDuration: meta.Duration,
		Author:        meta.Uploader,
		OriginalURL:   req.Url,
		CanonicalURL:  meta.WebpageURL,
	}

	// Fetch thumbnail
	if meta.Thumbnail != "" {
		p.log.Info("[yt-dlp] downloading thumbnail", "request", req.LogValue())
		if episode.ThumbnailFile, err = p.fetchThumbnail(ctx, req, meta.Thumbnail, thumbDir); err != nil {
			return nil, fmt.Errorf("failed to fetch thumbnail from youtube: %w", err)
		}
		p.log.Info("[yt-dlp] thumbnail downloaded", "request", req.LogValue())
	}

	// Fetch media
	p.log.Info("[yt-dlp] downloading media", "request", req.LogValue())
	if episode.MediaFile, episode.MediaSize, err = p.fetchMedia(ctx, req, mediaDir); err != nil {
		return nil, fmt.Errorf("failed to fetch media from youtube: %w", err)
	}

	p.log.Info("[yt-dlp] media downloaded", "request", req.LogValue())

	// Move thumbnail file to public directory
	if episode.ThumbnailFile != "" {
		if err = files.MoveFile(
			filepath.Join(thumbDir, episode.ThumbnailFile),
			filepath.Join(p.cfg.PublicDir, episode.ThumbnailFile),
		); err != nil {
			return nil, fmt.Errorf("failed to move thumbnail: %w", err)
		}
	}

	// Move media file to public directory
	if episode.MediaFile == "" {
		return nil, fmt.Errorf("media file name is empty")
	}
	if err = files.MoveFile(
		filepath.Join(mediaDir, episode.MediaFile),
		filepath.Join(p.cfg.PublicDir, episode.MediaFile),
	); err != nil {
		return nil, fmt.Errorf("failed to move media file: %w", err)
	}

	return episode, nil
}

func (p YtDlp) fetchMeta(ctx context.Context, req entities.Request, dir string) (*youtubeMeta, error) {
	cmd := exec.CommandContext(ctx, p.cfg.YtDlpPath,
		"--no-playlist", // Do not download playlists
		"-j",            // Dump JSON metadata
		"--no-warnings",
		"--skip-download",
		req.Url,
	)
	cmd.Dir = dir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp command failed: %w, stderr: %s", err, strings.TrimSpace(stderr.String()))
	}

	meta := &youtubeMeta{}
	if err := json.Unmarshal([]byte(stdout.String()), meta); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp json: %w", err)
	}

	if meta.Title == "" {
		meta.Title = meta.Uploader
	}
	if meta.Description == "" {
		meta.Description = "-"
	}

	return meta, nil
}

func (p YtDlp) fetchThumbnail(ctx context.Context, req entities.Request, thumbUrl, dir string) (string, error) {
	fileName := req.ID + ".jpg"
	cmd := exec.CommandContext(ctx, p.cfg.FFMpegPath,
		"-y",           // Overwrite output files without asking
		"-i", thumbUrl, // Input file
		"-vf", fmt.Sprintf(
			`crop=min(iw\,ih):min(iw\,ih):(iw-ow)/2:(ih-oh)/2,scale=%d:%d`,
			p.cfg.ThumbnailSize,
			p.cfg.ThumbnailSize,
		), // Crop to square and scale
		"-q:v", "2", // Quality level (1-31), lower is better
		fileName, // Output file
	)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg command failed: %w, output: %s", err, string(output))
	}
	return fileName, nil
}

func (p YtDlp) fetchMedia(ctx context.Context, req entities.Request, dir string) (string, int64, error) {
	fileName := req.ID + "." + string(req.DownloadFormat)
	cmd := exec.CommandContext(ctx, p.cfg.YtDlpPath,
		"--no-playlist",                              // Do not download playlists
		"-x",                                         // Extract audio
		"--audio-format", string(req.DownloadFormat), // Audio format
		"--audio-quality", req.DownloadQuality, // Audio quality
		"--no-warnings",     // Suppress warnings
		"--embed-thumbnail", // Embed thumbnail in the media file
		"--add-metadata",    // Add metadata to the media file
		"-o", fileName,      // Output file name
		"--force-overwrite", // Overwrite output files
		req.Url,
	)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", 0, fmt.Errorf("yt-dlp command failed: %w, output: %s", err, string(output))
	}

	// Get file size
	filePath := fmt.Sprintf("%s/%s", dir, fileName)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get file size: %w", err)
	}

	return fileName, fileInfo.Size(), nil
}

// youtubeMeta represents metadata fetched from yt-dlp.
type youtubeMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Duration    int64  `json:"duration"`
	Uploader    string `json:"uploader"`
	WebpageURL  string `json:"webpage_url"`
}

var mediaTypes = map[entities.DownloadFormat]entities.MediaType{
	entities.DownloadMp3: entities.MediaMp3,
}
