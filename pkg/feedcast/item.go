package feedcast

import (
	"strconv"
	"time"
)

// Item represents a podcast episode in the RSS feed.
type Item struct {
	xmlItem
}

// NewItem creates a new Item with the minimal required data
// as per Apple Podcasts specifications.
func NewItem(data ItemData) *Item {
	return &Item{
		xmlItem: xmlItem{
			Title: data.Title,
			Guid:  data.Guid,
			Enclosure: xmlEnclosure{
				URL:    data.Enclosure.URL,
				Length: data.Enclosure.Length,
				Type:   data.Enclosure.Type,
			},
		},
	}
}

// Validate checks if the item has the minimum required tags as per Apple Podcasts specifications.
// It returns an error if any required tag is missing or invalid.
//
// See https://help.apple.com/itc/podcasts_connect/#/itcb54353390
func (i *Item) Validate() error {
	return i.xmlItem.validate()
}

// WithPubDate sets the publication date for the episode.
// The date should be in RFC1123 format, e.g., "Mon, 02 Jan 2006 15:04:05 MST"
// which is equivalent to time.RFC1123Z in Go.
func (i *Item) WithPubDate(pubDate time.Time) *Item {
	i.xmlItem.PubDate = pubDate.Format(time.RFC1123Z)
	return i
}

// WithDescription sets the <description> tag containing one or more sentences
// describing your episode to potential listeners. You can specify up to 10,000 characters.
// You can use rich text formatting and some HTML (<p>, <ol>, <ul>, <li>, <a>)
// if wrapped in the <CDATA> tag.
func (i *Item) WithDescription(description string) *Item {
	i.xmlItem.Description = &xmlCDATA{Data: description}
	return i
}

// WithItunesDuration sets the <itunes:duration> tag containing the length of the episode in seconds.
func (i *Item) WithItunesDuration(duration int64) *Item {
	i.xmlItem.ItunesDuration = strconv.FormatInt(duration, 10)
	return i
}

// WithLink sets the <link> tag containing episode link URL.
// This is used when an episode has a corresponding webpage. Use the full URL
func (i *Item) WithLink(link string) *Item {
	i.xmlItem.Link = link
	return i
}

// WithItunesImage sets the <itunes:image> tag containing episode artwork URL.
// You should use this tag when you have a high quality, episode-specific image
// you would like listeners to see.
//
// Depending on their device, listeners see your episode artwork in varying sizes.
// Therefore, make sure your design is effective at both its original size and at thumbnail size.
// You should include a title, brand, or source name as part of your episode artwork.
// To avoid technical issues when you update your episode artwork, be sure to:
//   - Change the artwork file name and URL at the same time.
//   - Confirm your art does not contain an Alpha Channel.
//   - Verify the web server hosting your artwork allows HTTP head requests including Last Modified.
//   - Artwork must be a minimum size of 1400 x 1400 pixels
//     and a maximum size of 3000 x 3000 pixels, in JPEG or PNG format, 72 dpi,
//     with appropriate file extensions (.jpg, .png), and in the RGB colorspace.
//
// Make sure the file type in the URL matches the actual file type of the image file.
func (i *Item) WithItunesImage(href string) *Item {
	i.xmlItem.ItunesImage = &xmlItunesImage{Href: href}
	return i
}

// WithItunesExplicit sets the <itunes:explicit> tag containing parental advisory information.
// See Explicit for supported values.
func (i *Item) WithItunesExplicit(explicit Explicit) *Item {
	i.xmlItem.ItunesExplicit = explicit
	return i
}

// WithItunesTitle sets the <itunes:title> tag containing the episode title specific for Apple Podcasts.
// <itunes:title> is a string containing a clear concise name of your episode on Apple Podcasts.
// Don’t specify the episode number or season number in this tag.
// Instead, specify those details in the appropriate tags using WithItunesEpisode and WithItunesSeason.
// Also, don’t repeat the title of your show within your episode title.
// Separating episode and season number from the title makes it possible
// for Apple to easily index and order content from all shows.
func (i *Item) WithItunesTitle(title string) *Item {
	i.xmlItem.ItunesTitle = title
	return i
}

// WithItunesEpisode sets the <itunes:episode> tag containing a non-zero integer episode number.
// If all your episodes have numbers, and you would like them to be ordered based on them,
// use this tag for each one. Episode numbers are optional for episodic shows (TypeEpisodic),
// but are mandatory for serial shows (TypeSerial).
//
// If you are using your RSS feed to distribute a free version of an episode
// that is already available to Apple Podcasts paid subscribers, make sure
// the episode numbers are the same so you don’t have duplicate episodes appear on your show page.
// Learn more about how to set up your show for a subscription:
// https://podcasters.apple.com/support/set-up-your-show-for-a-subscription
func (i *Item) WithItunesEpisode(episode int) *Item {
	if episode > 0 {
		i.xmlItem.ItunesEpisode = strconv.Itoa(episode)
	}
	return i
}

// WithItunesSeason sets the <itunes:season> tag containing a non-zero integer season number.
// If an episode is within a season use this tag.
// To allow the season feature for shows containing a single season,
// if only one season exists in the RSS feed, Apple Podcasts doesn’t display a season number.
// When you add a second season to the RSS feed, Apple Podcasts displays the season numbers.
func (i *Item) WithItunesSeason(season int) *Item {
	if season > 0 {
		i.xmlItem.ItunesSeason = strconv.Itoa(season)
	}
	return i
}

// WithItunesEpisodeType sets the <itunes:episodeType> tag containing the episode type.
// If an episode is a trailer or bonus content, use this tag.
// See ItunesEpisodeType for supported values.
func (i *Item) WithItunesEpisodeType(episodeType ItunesEpisodeType) *Item {
	i.xmlItem.ItunesEpisodeType = episodeType
	return i
}

// WithPodcastTranscript sets the <podcast:transcript> tag containing link
// to the episode transcript in the Closed Caption format.
// You should use this tag when you have a valid transcript file available for users to read.
// Apple Podcasts will prefer VTT format over SRT format if multiple instances are included.
// See PodcastTranscriptType for supported values.
func (i *Item) WithPodcastTranscript(url string, transcriptType PodcastTranscriptType) *Item {
	i.xmlItem.PodcastTranscripts = append(i.xmlItem.PodcastTranscripts, xmlPodcastTranscript{
		Url:  url,
		Type: string(transcriptType),
	})
	return i
}

// WithItunesBlock sets the <itunes:block> tag for the episode.
// If you want an episode removed from the Apple directory, use this tag.
// For example, you might want to block a specific episode if you know that
// its content would otherwise cause the entire podcast to be removed from Apple Podcasts.
// See ItunesBlock for supported values.
func (i *Item) WithItunesBlock(block ItunesBlock) *Item {
	i.xmlItem.ItunesBlock = block
	return i
}

// WithItunesAuthor sets the <itunes:author> tag containing the author of the episode.
// If the episode has a different author than the podcast author,
// use this tag to specify the episode author.
func (i *Item) WithItunesAuthor(author string) *Item {
	i.xmlItem.ItunesAuthor = author
	return i
}

// WithItunesSummary sets the <itunes:summary> tag containing one or more sentences
// describing your episode to potential listeners. You can specify up to 4000 characters.
// You can use rich text formatting and some HTML (<p>, <ol>, <ul>, <li>, <a>)
// if wrapped in the <CDATA> tag.
func (i *Item) WithItunesSummary(summary string) *Item {
	i.xmlItem.ItunesSummary = &xmlCDATA{Data: summary}
	return i
}

// ItemData holds the data for the <item> element in the RSS feed.
// It includes the minimal set of required tags as per Apple Podcasts specifications
type ItemData struct {

	// Episode title: a clear, concise name for your episode.
	// Don’t specify the episode number or season number in this tag.
	// Instead, specify those details in the appropriate tags
	// using Item.WithItunesEpisode and Item.WithItunesSeason).
	// Also, don’t repeat the title of your show within your episode title.
	// Separating episode and season number from the title makes it possible
	// for Apple to easily index and order content from all shows.
	Title string

	// Episode content, file size, and file type information.
	Enclosure Enclosure

	// Episode’s globally unique identifier (Guid).
	//
	// It is very important that each episode have a unique Guid
	// and that it never changes, even if an episode’s metadata,
	// like title or enclosure URL, do change.
	// Globally unique identifiers (Guid) are case-sensitive strings.
	// If a Guid is not provided, an episode’s enclosure URL will be used instead.
	// If a Guid is not provided, make sure that an episode’s enclosure URL is unique and never changes.
	//
	// Failing to comply with these guidelines may result in duplicate episodes being shown to listeners,
	// inaccurate data in Analytics, and can cause issues with your podcasts’s
	// listing and chart placement in Apple Podcasts.
	Guid string
}

// Enclosure holds episode content, file size, and file type information.
//   - URL. The URL attribute points to your podcast media file.
//     Specify the full file extension within the URL attribute.
//     This determines whether or not content appears in the podcast directory.
//     Supported file formats include M4A, MP3, MOV, MP4, M4V, and PDF.
//   - Length. The length attribute is the file size in bytes.
//   - Type. The type attribute provides the correct category
//     for the type of file you are using.
//     See EnclosureType for supported values.
type Enclosure struct {

	// The URL attribute points to your podcast media file.
	// Specify the full file extension within the URL attribute.
	// This determines whether or not content appears in the podcast directory.
	// Supported file formats include M4A, MP3, MOV, MP4, M4V, and PDF.
	URL string

	// The Length attribute is the file size in bytes.
	Length int64

	// The Type attribute provides the correct category
	// for the type of file you are using.
	// See EnclosureType for supported values.
	Type EnclosureType
}

// NewEnclosure creates a new Enclosure with the given parameters:
//   - The URL attribute points to your podcast media file.
//     Specify the full file extension within the URL attribute.
//     This determines whether or not content appears in the podcast directory.
//     Supported file formats include M4A, MP3, MOV, MP4, M4V, and PDF.
//   - The length attribute is the file size in bytes.
//   - The type attribute provides the correct category
//     for the type of file you are using.
//     See EnclosureType for supported values.
func NewEnclosure(url string, length int64, enclosureType EnclosureType) Enclosure {
	return Enclosure{
		URL:    url,
		Length: length,
		Type:   enclosureType,
	}
}
