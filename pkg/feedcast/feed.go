package feedcast

import (
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

// Feed represents an RSS podcast feed.
// It contains the channel information and a list of items (episodes).
type Feed struct {
	xmlDoc
}

// NewFeed creates a new Feed instance with the provided channel data and categories.
// It initializes the feed with minimum required channel tags as per Apple Podcasts specifications.
func NewFeed(channel FeedData) *Feed {
	cat := make([]xmlItunesCategory, len(channel.Categories))
	for i, c := range channel.Categories {
		sub := make([]xmlItunesCategory, len(c.Subcategories))
		for j, s := range c.Subcategories {
			sub[j] = xmlItunesCategory{Text: s}
		}
		cat[i] = xmlItunesCategory{
			Text:             c.Text,
			ItunesCategories: sub,
		}
	}

	return &Feed{
		xmlDoc: xmlDoc{
			Version:   xmlRssVersion,
			ItunesNS:  xmlItunesNS,
			ContentNS: xmlContentNS,
			PodcastNS: xmlPodcastNS,
			Channel: xmlChannel{
				Title:          channel.Title,
				Description:    xmlCDATA{Data: channel.Description},
				ItunesImage:    xmlItunesImage{Href: channel.Image},
				Language:       channel.Language,
				ItunesExplicit: channel.Explicit,
				Items:          []xmlItem{},
				ItunesCategory: cat,
			},
		},
	}
}

// Validate checks if the feed has the minimum required tags as per Apple Podcasts specifications.
// It returns an error if any required tag is missing or invalid.
//
// See https://help.apple.com/itc/podcasts_connect/#/itcb54353390
func (f *Feed) Validate() error {
	return f.xmlDoc.validate()
}

// AddItem adds an episode item to the feed.
func (f *Feed) AddItem(item *Item) {
	f.xmlDoc.Channel.Items = append(f.xmlDoc.Channel.Items, item.xmlItem)
}

func (f *Feed) Encode(w io.Writer) error {
	if err := f.Validate(); err != nil {
		return fmt.Errorf("feed validation failed: %w", err)
	}
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return fmt.Errorf("failed to write xml header: %w", err)
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(f.xmlDoc); err != nil {
		return fmt.Errorf("failed to encode feed: %w", err)
	}
	return nil
}

// WithAuthor sets the <itunes:author> tag of the feed.
// Author is the group responsible for creating the show.
//
// Show author most often refers to the parent company or network of a podcast,
// but it can also be used to identify the host(s) if none exists.
//
// Author information is especially useful if a company or organization publishes multiple podcasts.
func (f *Feed) WithAuthor(name string) *Feed {
	f.xmlDoc.Channel.ItunesAuthor = name
	return f
}

// WithLink sets the <link> tag of the feed of the website associated with a podcast. Use the full URL.
//
// Typically, a home page for a podcast or a dedicated portion of a larger website.
func (f *Feed) WithLink(link string) *Feed {
	f.xmlDoc.Channel.Link = link
	return f
}

// WithPubDate sets the <pubDate> tag of the feed.
// It indicates the publication date and time for the content of the feed.
//
// The value for this tag is typically the publication date of the most recent episode in the feed.
//
// The date should be in RFC 2822 format (e.g., "Sat, 01 Apr 2023 19:00:00 GMT")
// which is the same as time.RFC1123Z in Go.
func (f *Feed) WithPubDate(pubDate time.Time) *Feed {
	f.xmlDoc.Channel.PubDate = pubDate.Format(time.RFC1123Z)
	return f
}

// WithLastBuildDate sets the <lastBuildDate> tag of the feed.
// It indicates the last time the content of the feed was modified.
//
// The date should be in RFC 2822 format (e.g., "Sat, 01 Apr 2023 19:00:00 GMT")
// which is the same as time.RFC1123Z in Go.
func (f *Feed) WithLastBuildDate(lastBuildDate time.Time) *Feed {
	f.xmlDoc.Channel.LastBuildDate = lastBuildDate.Format(time.RFC1123Z)
	return f
}

// WithItunesTitle sets the <itunes:title> tag specific for Apple Podcasts.
//
// <itunes:title> is a string containing a clear concise name of your show on Apple Podcasts.
// Do not include episode or season number in the title. There are dedicated tags for that information.
// See WithItunesEpisode and WithItunesSeason.
func (f *Feed) WithItunesTitle(title string) *Feed {
	f.xmlDoc.Channel.ItunesTitle = title
	return f
}

// WithItunesType sets the <itunes:type> tag of type of show.
// If your show is Serial you must use this tag.
// See ItunesType for possible values.
func (f *Feed) WithItunesType(showType ItunesType) *Feed {
	f.xmlDoc.Channel.ItunesType = showType
	return f
}

// WithCopyright sets the <copyright> tag of the feed.
// If your show is copyrighted you should use this tag.
func (f *Feed) WithCopyright(copyright string) *Feed {
	f.xmlDoc.Channel.Copyright = copyright
	return f
}

// WithItunesNewFeedURL sets the <itunes:new-feed-url> tag of the feed.
// Use this tag if you are changing the URL of your podcast feed.
// This tag should contain the new URL of your podcast feed.
//
// You should maintain your old feed until you have migrated your existing followers.
// Learn how to update your podcast RSS feed URL:
// https://podcasters.apple.com/support/change-the-rss-feed-url
//
// Note: The <itunes:new-feed-url> tag reports new feed URLs
// to Apple Podcasts and isn’t displayed in Apple Podcasts.
func (f *Feed) WithItunesNewFeedURL(newFeedURL string) *Feed {
	f.xmlDoc.Channel.ItunesNewFeedURL = newFeedURL
	return f
}

// WithItunesBlock sets the <itunes:block> tag of the feed.
// Use this tag if you want to block your podcast from appearing in Apple Podcasts.
// See ItunesBlock for possible values.
func (f *Feed) WithItunesBlock(block ItunesBlock) *Feed {
	f.xmlDoc.Channel.ItunesBlock = block
	return f
}

// WithItunesComplete sets the <itunes:complete> tag of the feed.
// Use this tag if your podcast is complete and no new episodes will be added.
// See ItunesComplete for possible values.
func (f *Feed) WithItunesComplete(complete ItunesComplete) *Feed {
	f.xmlDoc.Channel.ItunesComplete = complete
	return f
}

// WithGenerator sets the <generator> tag of the feed.
// This tag is typically used to identify the software that generated the feed.
// Hosting providers use this tag to identify themselves as the creator of an RSS feed.
func (f *Feed) WithGenerator(generator string) *Feed {
	f.xmlDoc.Channel.Generator = generator
	return f
}

// WithItunesSummary sets the <itunes:summary> tag of the feed.
// This tag is similar to the <description> tag but is specific to Apple Podcasts.
// It provides a summary of the podcast show.
func (f *Feed) WithItunesSummary(summary string) *Feed {
	f.xmlDoc.Channel.ItunesSummary = &xmlCDATA{Data: summary}
	return f
}

// WithItunesKeywords sets the <itunes:keywords> tag of the feed.
// This tag is a comma-separated list of keywords that describe your podcast.
// Keywords help users find your podcast when they search in Apple Podcasts.
func (f *Feed) WithItunesKeywords(keywords string) *Feed {
	f.xmlDoc.Channel.ItunesKeywords = keywords
	return f
}

// WithItunesOwner sets the <itunes:owner> tag of the feed.
// This tag contains information about the owner of the podcast, including name and email.
func (f *Feed) WithItunesOwner(name, email string) *Feed {
	f.xmlDoc.Channel.ItunesOwner = &xmlItunesOwner{
		Name:  name,
		Email: email,
	}
	return f
}

// FeedData holds the data for the <channel> element in the RSS feed.
// It includes the minimal set of required tags as per Apple Podcasts specifications.
type FeedData struct {
	// The show title
	Title string

	// The show description. Where description is text containing
	// one or more sentences describing your podcast to potential listeners.
	// The maximum amount of text allowed for this tag is 4000 bytes.
	Description string

	// The artwork for the show. Specify your show artwork by providing a URL linking to it.
	// Depending on their device, subscribers see your podcast artwork in varying sizes.
	// Therefore, make sure your design is effective at both its original size and at thumbnail size.
	// You should include a show title, brand, or source name as part of your podcast artwork.
	// Here are additional marketing best practices.
	// For examples of podcast artwork, see the Top Podcasts chart.
	// To avoid technical issues when you update your podcast artwork, be sure to:
	//   - Change the artwork file name and URL at the same time
	//   - Make sure the file type in the URL matches the actual file type of the image file.
	//   - Verify the web server hosting your artwork allows HTTP head requests including Last Modified.
	//
	// Artwork must be a minimum size of 1400 x 1400 pixels and a maximum size of 3000 x 3000 pixels,
	// in JPEG or PNG format, 72 dpi, with appropriate file extensions (.jpg, .png), and in the RGB colorspace.
	// Confirm your art does not contain an Alpha Channel.
	// These requirements are different from the standard RSS image tag specifications.
	Image string

	// The language spoken on the show.
	// Because Apple Podcasts is available in territories around the world,
	// it is critical to specify the language of a podcast.
	// Apple Podcasts only supports values from the ISO 639 list.
	//
	// https://www.loc.gov/standards/iso639-2/php/code_list.php
	//
	// Invalid language codes will cause your feed to fail Apple validation.
	Language string

	// The parental advisory information.
	// The explicit value can be one of the following:
	//   - ExplicitTrue. If you specify true, indicating the presence of explicit content,
	//     Apple Podcasts displays an Explicit parental advisory graphic for your podcast.
	//     Podcasts containing explicit material aren’t available in some Apple Podcasts territories.
	//   - ExplicitFalse. If you specify false, indicating that your podcast doesn’t contain
	//     explicit language or adult content, Apple Podcasts displays a Clean parental advisory graphic
	//     for your podcast.
	Explicit Explicit

	// The podcast categories. At least one category is required.
	// See Category for possible values.
	Categories []Category
}
