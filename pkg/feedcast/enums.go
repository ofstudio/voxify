package feedcast

// ItunesType represents the type of show.
// If your show is Serial you must use this tag.
// Its values can be one of the following:
//   - TypeEpisodic. Specify episodic when episodes are intended to be consumed
//     without any specific order. Apple Podcasts will present newest episodes first and display
//     the publish date (required) of each episode. If organized into seasons,
//     the newest season will be presented first - otherwise, episodes will be grouped
//     by year published, newest first. For new subscribers, Apple Podcasts adds the newest,
//     most recent episode in their Library.
//   - Serial. Specify serial when episodes are intended to be consumed in sequential order.
//     Apple Podcasts will present the oldest episodes first and display the episode
//     numbers (required) of each episode. If organized into seasons, the newest season
//     will be presented first and <itunes:episode> numbers must be given for each episode.
//
// Each show type has different behavior for automatic downloads.
// See: https://podcasters.apple.com/support/1662-automatic-downloads-on-apple-podcasts
type ItunesType string

const (
	TypeNotSet   ItunesType = ""
	TypeEpisodic ItunesType = "episodic"
	TypeSerial   ItunesType = "serial"
)

// ItunesBlock represent podcast show or hide status.
// If you want your show removed from the Apple directory, use this tag.
// Specifying the <itunes:block> tag with a BlockYes value,
// prevents the entire podcast from appearing in Apple Podcasts.
// Specifying any other value has no effect.
type ItunesBlock string

const (
	BlockNotSet ItunesBlock = ""
	BlockYes    ItunesBlock = "Yes"
)

// ItunesComplete represents the podcast  update status.
// If you will never publish another episode to your show, use this tag.
// Specifying the <itunes:complete> tag with a CompleteYes value
// indicates that a podcast is complete and you will not post
// any more episodes in the future.
// Specifying any other value has no effect.
type ItunesComplete string

const (
	CompleteNotSet ItunesComplete = ""
	CompleteYes    ItunesComplete = "Yes"
)

// EnclosureType provides the correct category for the type of file you are using.
// The type values for the supported file formats are:
//   - M4a for audio/x-m4a
//   - Mp3 for audio/mpeg
//   - Mov for video/quicktime
//   - Mp4 for video/mp4
//   - M4v for video/x-m4v
//   - Pdf for application/pdf
type EnclosureType string

const (
	M4a EnclosureType = "audio/x-m4a"
	Mp3 EnclosureType = "audio/mpeg"
	Mov EnclosureType = "video/quicktime"
	Mp4 EnclosureType = "video/mp4"
	M4v EnclosureType = "video/x-m4v"
	Pdf EnclosureType = "application/pdf."
)

// Explicit is the parental advisory information.
// The explicit value can be one of the following:
//   - ExplicitTrue. If you specify true, indicating the presence of explicit content,
//     Apple Podcasts displays an Explicit parental advisory graphic for your podcast.
//     Podcasts containing explicit material aren’t available in some Apple Podcasts territories.
//   - ExplicitFalse. If you specify false, indicating that your podcast doesn’t contain
//     explicit language or adult content, Apple Podcasts displays a Clean parental advisory graphic
//     for your podcast.
type Explicit string

const (
	ExplicitNotSet Explicit = ""
	ExplicitTrue   Explicit = "true"
	ExplicitFalse  Explicit = "false"
)

// ItunesEpisodeType is the episode type. If an episode is a trailer or bonus content, use this tag.
// Where the episodeType value can be one of the following:
//   - EpisodeFull. Specify full when you are submitting the complete content of your show.
//   - EpisodeTrailer. Specify trailer when you are submitting a short, promotional
//     piece of content that represents a preview of your current show.
//   - EpisodeBonus. Specify bonus when you are submitting extra content for your show
//     (for example, behind the scenes information or interviews with the cast) or cross-promotional
//     content for another show.
//
// The rules for using trailer and bonus tags depend on whether the <itunes:season>
// and <itunes:episode> tags have values:
//
// # Trailer
//   - No season or episode number: a show trailer
//   - A season number and no episode number: a season trailer. (Note: an episode trailer should have a different <guid> than the actual episode)
//   - Episode number and optionally a season number: an episode trailer/teaser, later replaced with the actual episode
//
// # Bonus
//
//   - No season or episode number: a show bonus
//   - A season number: a season bonus
//   - Episode number and optionally a season number: a bonus episode related to a specific episode
type ItunesEpisodeType string

const (
	EpisodeFull    ItunesEpisodeType = "full"
	EpisodeTrailer ItunesEpisodeType = "trailer"
	EpisodeBonus   ItunesEpisodeType = "bonus"
)

// PodcastTranscriptType represents the format of the transcript file.
// Apple Podcasts will prefer VTT format over SRT format if multiple instances are included.
// A valid type attribute is required. Accepted types include:
//   - TranscriptVtt for text/vtt
//   - TranscriptSrt for application/srt,
//   - TranscriptSubrip for application/x-subrip.
type PodcastTranscriptType string

const (
	TranscriptVtt    PodcastTranscriptType = "text/vtt"
	TranscriptSrt    PodcastTranscriptType = "application/srt"
	TranscriptSubrip PodcastTranscriptType = " application/x-subrip"
)
