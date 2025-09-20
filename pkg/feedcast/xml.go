package feedcast

import (
	"encoding/xml"
	"errors"
	"fmt"
)

const (
	xmlItunesNS   = "http://www.itunes.com/dtds/podcast-1.0.dtd"
	xmlContentNS  = "http://purl.org/rss/1.0/modules/content/"
	xmlPodcastNS  = "https://podcastindex.org/namespace/1.0"
	xmlRssVersion = "2.0"
)

// xmlDoc represents the root <rss> element in the RSS feed.
type xmlDoc struct {
	XMLName   xml.Name   `xml:"rss"`
	Version   string     `xml:"version,attr"`
	ItunesNS  string     `xml:"xmlns:itunes,attr"`
	ContentNS string     `xml:"xmlns:content,attr"`
	PodcastNS string     `xml:"xmlns:podcast,attr"`
	Channel   xmlChannel `xml:"channel"`
}

func (d *xmlDoc) validate() error {
	if d.Version != xmlRssVersion {
		return errors.New("feed version mismatch")
	}
	if d.ItunesNS != xmlItunesNS {
		return errors.New("itunes namespace mismatch")
	}
	if d.ContentNS != xmlContentNS {
		return errors.New("content namespace mismatch")
	}
	if d.PodcastNS != xmlPodcastNS {
		return errors.New("podcast namespace mismatch")
	}
	return d.Channel.validate()
}

// xmlChannel represents the <channel> element in the RSS feed.
type xmlChannel struct {
	// Required tags
	Title          string              `xml:"title"`
	Description    xmlCDATA            `xml:"description"`
	ItunesImage    xmlItunesImage      `xml:"itunes:image"`
	Language       string              `xml:"language"`
	ItunesExplicit Explicit            `xml:"itunes:explicit"`
	ItunesCategory []xmlItunesCategory `xml:"itunes:category"`

	// Recommended tags
	ItunesAuthor  string `xml:"itunes:author,omitempty"`
	Link          string `xml:"link,omitempty"`
	PubDate       string `xml:"pubDate,omitempty"`
	LastBuildDate string `xml:"lastBuildDate,omitempty"`

	// Situational tags
	ItunesTitle      string          `xml:"itunes:title,omitempty"`
	ItunesType       ItunesType      `xml:"itunes:type,omitempty"`
	Copyright        string          `xml:"copyright,omitempty"`
	ItunesNewFeedURL string          `xml:"itunes:new-feed-url,omitempty"`
	ItunesBlock      ItunesBlock     `xml:"itunes:block,omitempty"`
	ItunesComplete   ItunesComplete  `xml:"itunes:complete,omitempty"`
	Generator        string          `xml:"generator,omitempty"`
	ItunesSummary    *xmlCDATA       `xml:"itunes:summary,omitempty"`
	ItunesKeywords   string          `xml:"itunes:keywords,omitempty"`
	ItunesOwner      *xmlItunesOwner `xml:"itunes:owner,omitempty"`

	// Items (episodes)
	Items []xmlItem `xml:"item"`
}

func (c *xmlChannel) validate() error {
	if c.Title == "" {
		return errors.New("channel title is required")
	}
	if c.Description.Data == "" {
		return errors.New("channel description is required")
	}
	if c.ItunesImage.Href == "" {
		return errors.New("channel itunes:image is required")
	}
	if c.Language == "" {
		return errors.New("channel language is required")
	}
	if c.ItunesExplicit != ExplicitTrue && c.ItunesExplicit != ExplicitFalse {
		return errors.New("channel itunes:explicit must be either 'true' or 'false'")
	}
	if len(c.ItunesCategory) == 0 {
		return errors.New("at least one itunes:category is required")
	}
	if len(c.Items) == 0 {
		return errors.New("at least one channel item is required")
	}
	for i, item := range c.Items {
		if err := item.validate(); err != nil {
			return fmt.Errorf("invalid item %d: %w", i, err)
		}
	}
	return nil
}

// xmlItem represents the <item> element in the RSS feed.
type xmlItem struct {
	// Required tags
	Title     string       `xml:"title"`
	Enclosure xmlEnclosure `xml:"enclosure"`
	Guid      string       `xml:"guid"`

	// Recommended tags
	PubDate        string          `xml:"pubDate,omitempty"`
	Description    *xmlCDATA       `xml:"description,omitempty"`
	ItunesDuration string          `xml:"itunes:duration,omitempty"`
	Link           string          `xml:"link,omitempty"`
	ItunesImage    *xmlItunesImage `xml:"itunes:image,omitempty"`
	ItunesExplicit Explicit        `xml:"itunes:explicit,omitempty"`

	// Situational tags
	ItunesTitle        string                 `xml:"itunes:title,omitempty"`
	ItunesEpisode      string                 `xml:"itunes:episode,omitempty"`
	ItunesSeason       string                 `xml:"itunes:season,omitempty"`
	ItunesEpisodeType  ItunesEpisodeType      `xml:"itunes:episodeType,omitempty"`
	PodcastTranscripts []xmlPodcastTranscript `xml:"podcast:transcript,omitempty"`
	ItunesBlock        ItunesBlock            `xml:"itunes:block,omitempty"`
	ItunesAuthor       string                 `xml:"itunes:author,omitempty"`
	ItunesSummary      *xmlCDATA              `xml:"itunes:summary,omitempty"`
}

func (i *xmlItem) validate() error {
	if i.Title == "" {
		return errors.New("item title is required")
	}
	if i.Enclosure.URL == "" {
		return errors.New("item enclosure url is required")
	}
	if i.Enclosure.Length <= 0 {
		return errors.New("item enclosure length must be greater than zero")
	}
	if i.Enclosure.Type == "" {
		return errors.New("item enclosure type is required")
	}
	if i.Guid == "" {
		return errors.New("item guid is required")
	}
	if i.ItunesExplicit != "" && i.ItunesExplicit != ExplicitTrue && i.ItunesExplicit != ExplicitFalse {
		return errors.New("item itunes:explicit must be either 'true' or 'false'")
	}
	return nil
}

// xmlItunesImage represents the <itunes:image> element in the RSS feed.
type xmlItunesImage struct {
	XMLName xml.Name `xml:"itunes:image"`
	Href    string   `xml:"href,attr"`
}

// xmlItunesCategory represents the <itunes:category> element in the RSS feed.
type xmlItunesCategory struct {
	XMLName          xml.Name            `xml:"itunes:category"`
	Text             string              `xml:"text,attr"`
	ItunesCategories []xmlItunesCategory `xml:"itunes:category,omitempty"`
}

// xmlEnclosure represents the <enclosure> element in the RSS feed.
type xmlEnclosure struct {
	URL    string        `xml:"url,attr"`
	Length int64         `xml:"length,attr"`
	Type   EnclosureType `xml:"type,attr"`
}

// xmlItunesOwner represents the <itunes:owner> element in the RSS feed.
type xmlItunesOwner struct {
	Name  string `xml:"itunes:name"`
	Email string `xml:"itunes:email"`
}

// xmlPodcastTranscript represents the <podcast:transcript> element in the RSS feed.
type xmlPodcastTranscript struct {
	Url  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

// xmlCDATA is a custom type to handle CDATA sections in XML.
type xmlCDATA struct {
	Data string `xml:",cdata"`
}
