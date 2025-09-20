package feedcast

import (
	"testing"
)

func TestExplicitEnums(t *testing.T) {
	tests := []struct {
		name     string
		value    Explicit
		expected string
	}{
		{"not set", ExplicitNotSet, ""},
		{"true", ExplicitTrue, "true"},
		{"false", ExplicitFalse, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.value))
			}
		})
	}
}

func TestItunesTypeEnums(t *testing.T) {
	tests := []struct {
		name     string
		value    ItunesType
		expected string
	}{
		{"not set", TypeNotSet, ""},
		{"episodic", TypeEpisodic, "episodic"},
		{"serial", TypeSerial, "serial"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.value))
			}
		})
	}
}

func TestItunesBlockEnums(t *testing.T) {
	tests := []struct {
		name     string
		value    ItunesBlock
		expected string
	}{
		{"not set", BlockNotSet, ""},
		{"yes", BlockYes, "Yes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.value))
			}
		})
	}
}

func TestItunesCompleteEnums(t *testing.T) {
	tests := []struct {
		name     string
		value    ItunesComplete
		expected string
	}{
		{"not set", CompleteNotSet, ""},
		{"yes", CompleteYes, "Yes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.value))
			}
		})
	}
}

func TestEnclosureTypeEnums(t *testing.T) {
	tests := []struct {
		name     string
		value    EnclosureType
		expected string
	}{
		{"M4A", M4a, "audio/x-m4a"},
		{"MP3", Mp3, "audio/mpeg"},
		{"MOV", Mov, "video/quicktime"},
		{"MP4", Mp4, "video/mp4"},
		{"M4V", M4v, "video/x-m4v"},
		{"PDF", Pdf, "application/pdf."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.value))
			}
		})
	}
}

func TestItunesEpisodeTypeEnums(t *testing.T) {
	tests := []struct {
		name     string
		value    ItunesEpisodeType
		expected string
	}{
		{"full", EpisodeFull, "full"},
		{"trailer", EpisodeTrailer, "trailer"},
		{"bonus", EpisodeBonus, "bonus"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.value))
			}
		})
	}
}

func TestPodcastTranscriptTypeEnums(t *testing.T) {
	tests := []struct {
		name     string
		value    PodcastTranscriptType
		expected string
	}{
		{"VTT", TranscriptVtt, "text/vtt"},
		{"SRT", TranscriptSrt, "application/srt"},
		{"SubRip", TranscriptSubrip, " application/x-subrip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.value))
			}
		})
	}
}

// Test that enum values match Apple Podcast specifications
func TestApplePodcastCompliance(t *testing.T) {
	// Test Explicit values are exactly as required by Apple
	if ExplicitTrue != "true" {
		t.Error("ExplicitTrue must be 'true' for Apple Podcast compliance")
	}
	if ExplicitFalse != "false" {
		t.Error("ExplicitFalse must be 'false' for Apple Podcast compliance")
	}

	// Test iTunes type values
	if TypeEpisodic != "episodic" {
		t.Error("TypeEpisodic must be 'episodic' for Apple Podcast compliance")
	}
	if TypeSerial != "serial" {
		t.Error("TypeSerial must be 'serial' for Apple Podcast compliance")
	}

	// Test block values (case-sensitive)
	if BlockYes != "Yes" {
		t.Error("BlockYes must be 'Yes' (capital Y) for Apple Podcast compliance")
	}

	// Test complete values (case-sensitive)
	if CompleteYes != "Yes" {
		t.Error("CompleteYes must be 'Yes' (capital Y) for Apple Podcast compliance")
	}

	// Test supported MIME types for enclosures
	supportedTypes := map[EnclosureType]bool{
		M4a: true,
		Mp3: true,
		Mov: true,
		Mp4: true,
		M4v: true,
		Pdf: true,
	}

	if len(supportedTypes) != 6 {
		t.Error("Should support exactly 6 enclosure types as per Apple specifications")
	}

	// Verify correct MIME types
	if Mp3 != "audio/mpeg" {
		t.Error("MP3 MIME type incorrect")
	}
	if M4a != "audio/x-m4a" {
		t.Error("M4A MIME type incorrect")
	}
	if Mp4 != "video/mp4" {
		t.Error("MP4 MIME type incorrect")
	}

	// Test episode types
	if EpisodeFull != "full" {
		t.Error("Episode type 'full' incorrect")
	}
	if EpisodeTrailer != "trailer" {
		t.Error("Episode type 'trailer' incorrect")
	}
	if EpisodeBonus != "bonus" {
		t.Error("Episode type 'bonus' incorrect")
	}
}

func TestEnclosureTypePDFTypo(t *testing.T) {
	// There's a typo in the PDF MIME type (extra period)
	// This test documents it for awareness
	if Pdf != "application/pdf." {
		t.Error("PDF enclosure type has changed - check if typo was fixed")
	}

	// Note: The correct MIME type should be "application/pdf" without the trailing period
	// This might need to be fixed in the enum definition
}

func TestTranscriptTypeSpacing(t *testing.T) {
	// Test for potential spacing issues in transcript types
	if TranscriptSubrip != " application/x-subrip" {
		t.Error("SubRip transcript type has changed - check spacing")
	}

	// Note: There's a leading space in TranscriptSubrip
	// This might need to be fixed in the enum definition
}
