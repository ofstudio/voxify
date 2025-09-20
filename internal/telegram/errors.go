package telegram

import (
	"errors"
	"fmt"

	"github.com/ofstudio/voxify/internal/locales"
	"github.com/ofstudio/voxify/internal/services"
)

var (
	errProcessorBusy = errors.New("processor is busy")
)

// msgErr returns message to be sent to user based on the error
func msgErr(err error) string {
	if err == nil {
		return fmt.Sprintf(locales.MsgSomethingWentWrongWithCode, 0)
	}

	// Handle specific known error
	if errors.Is(err, errProcessorBusy) {
		return locales.MsgDownloadBusy
	}

	// Handle business logic errors with codes
	var e services.Error
	if errors.As(err, &e) {
		switch e.Code {
		case 101:
			return locales.MsgNoMatchingPlatform
		case 102:
			return locales.MsgDownloadFailed
		case 103:
			return locales.MsgEpisodeInProgress
		case 104:
			return locales.MsgEpisodeExists
		case 105:
			return locales.MsgProcessInterrupted
		case 106:
			return locales.MsgEmptyFeed
		case 107:
			return locales.MsgInvalidRequest
		default:
			return fmt.Sprintf(locales.MsgSomethingWentWrongWithCode, e.Code)
		}
	}
	return locales.MsgSomethingWentWrong
}
