package entities

import (
	"time"

	"github.com/ofstudio/voxify/pkg/feedcast"
)

// Feed contains information about podcast feed
type Feed struct {
	Title         string         // Show title
	Description   string         // Show description
	Summary       string         // Show summary
	Language      string         // Lang code
	Categories    []FeedCategory // Podcast categories
	Keywords      string         // Podcast keywords comma separated
	Author        string         // Podcast author name
	Owner         *FeedOwner     // Podcast owner info
	Copyright     string         // Show copyright info if any
	Explicit      bool           // Is the show explicit
	FeedType      FeedType       // The type of the show
	FeedCompleted bool           // Is the show completed
	FeedBlocked   bool           // Is the show blocked in Apple Podcasts
	WebsiteLink   string         // Website link
	RSSLink       string         // RSS feed link
	ImageUrl      string         // Cover image link
	Generator     string         // Feed generator software
	PubDate       time.Time      // Last published date. Zero time if no episodes
	EpisodeCount  int            // Number of episodes in the feed
}

// FeedOwner contains information about podcast owner
type FeedOwner struct {
	Name  string // Owner name
	Email string // Owner email
}

// FeedCategory contains information about podcast category
type FeedCategory = feedcast.Category

// FeedType represents the type of show.
// If your show is Serial you must use this tag.
// See [feedcast.ItunesType] for possible values.
type FeedType = feedcast.ItunesType

const (
	FeedTypeNotSet   = feedcast.TypeNotSet
	FeedTypeEpisodic = feedcast.TypeEpisodic
	FeedTypeSerial   = feedcast.TypeSerial
)
