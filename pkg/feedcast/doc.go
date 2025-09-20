// Package feedcast provides a comprehensive toolkit for generating RSS podcast feeds
// that comply with Apple Podcast specifications and modern podcast standards.
//
// The package enables creation of podcast RSS feeds with full support for:
//   - Apple Podcast required and optional tags
//   - Episode metadata (titles, descriptions, artwork, transcripts)
//   - Podcast categories and subcategories
//   - Serial and episodic show types
//   - Rich content with HTML and CDATA support
//   - Comprehensive validation against Apple's requirements
//
// # Basic Usage
//
// Creating a simple podcast feed:
//
//	package main
//
//	import (
//		"os"
//		"time"
//		"github.com/ofstudio/voxify/pkg/feedcast"
//	)
//
//	func main() {
//		// Create channel data
//		channelData := feedcast.FeedData{
//			Title:       "My Tech Podcast",
//			Description: "Weekly discussions about technology trends and innovations",
//			Image:       "https://example.com/podcast-artwork.jpg",
//			Language:    "en",
//			Explicit:    feedcast.ExplicitFalse,
//			Categories: []feedcast.Category{
//				feedcast.NewCategory("Technology"),
//				feedcast.NewCategory("Business", "Entrepreneurship"),
//			},
//		}
//
//		// Create feed with optional metadata
//		feed := feedcast.NewFeed(channelData).
//			WithAuthor("Tech Media Network").
//			WithLink("https://mytechpodcast.com").
//			WithItunesType(feedcast.TypeEpisodic).
//			WithItunesOwner("John Doe", "john@example.com")
//
//		// Add an episode
//		episodeData := feedcast.ItemData{
//			Title: "The Future of AI",
//			Guid:  "episode-001",
//			Enclosure: feedcast.Enclosure{
//				URL:    "https://example.com/episodes/episode-001.mp3",
//				Length: 25600000, // bytes
//				Type:   feedcast.Mp3,
//			},
//		}
//
//		episode := feedcast.NewItem(episodeData).
//			WithPubDate(time.Now()).
//			WithDescription("Deep dive into artificial intelligence trends").
//			WithItunesDuration(3600). // 1 hour in seconds
//			WithItunesEpisode(1).
//			WithItunesSeason(1).
//			WithPodcastTranscript("https://example.com/transcripts/001.vtt", feedcast.TranscriptVtt)
//
//		feed.AddItem(episode)
//
//		// Generate RSS XML
//		if err := feed.Validate(); err != nil {
//			panic(err)
//		}
//
//		feed.Encode(os.Stdout)
//	}
//
// # Validation and Compliance
//
// The package automatically validates feeds against Apple Podcast requirements:
//   - Required channel tags (title, description, image, language, explicit, categories)
//   - Required episode tags (title, Guid, enclosure)
//   - Proper MIME types for audio/video files
//   - Valid category selections
//   - Correct namespace declarations and RSS structure
//
// # References
//
// This package implements specifications from:
//   - Apple Podcast Requirements: https://podcasters.apple.com/support/823-podcast-requirements
//   - Apple Podcast Connect Guide: https://help.apple.com/itc/podcasts_connect/#/itcb54353390
//   - The "podcast" Namespace: https://github.com/Podcastindex-org/podcast-namespace
package feedcast
