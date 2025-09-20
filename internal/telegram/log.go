package telegram

import (
	"log/slog"

	"github.com/go-telegram/bot/models"
)

// logUpdate returns a slog.Value for models.Update for structured logging
func logUpdate(u *models.Update) slog.Value {
	if u == nil {
		return slog.Value{}
	}

	attrs := []slog.Attr{
		slog.Int64("id", u.ID),
	}

	if u.Message != nil {
		attrs = append(attrs, slog.Any("message", logMessage(u.Message)))
	}
	if u.EditedMessage != nil {
		attrs = append(attrs, slog.Any("edited_message", logMessage(u.EditedMessage)))
	}
	if u.ChannelPost != nil {
		attrs = append(attrs, slog.Any("channel_post", logMessage(u.ChannelPost)))
	}
	if u.EditedChannelPost != nil {
		attrs = append(attrs, slog.Any("edited_channel_post", logMessage(u.EditedChannelPost)))
	}
	if u.CallbackQuery != nil {
		attrs = append(attrs, slog.String("callback_query_id", u.CallbackQuery.ID))
	}
	if u.InlineQuery != nil {
		attrs = append(attrs, slog.String("inline_query_id", u.InlineQuery.ID))
	}

	return slog.GroupValue(attrs...)
}

// logMessage returns a slog.Value for models.Message for structured logging
func logMessage(m *models.Message) slog.Value {
	if m == nil {
		return slog.Value{}
	}

	attrs := []slog.Attr{
		slog.Int("id", m.ID),
		slog.Int64("date", int64(m.Date)),
		slog.Any("chat", logChat(&m.Chat)),
	}

	if m.From != nil {
		attrs = append(attrs, slog.Any("from", LogUser(m.From)))
	}
	if m.ReplyToMessage != nil {
		attrs = append(attrs, slog.Int("reply_to_message_id", m.ReplyToMessage.ID))
	}
	if m.ForwardOrigin != nil {
		attrs = append(attrs, slog.String("forward_origin_type", string(m.ForwardOrigin.Type)))
	}
	if m.Document != nil {
		attrs = append(attrs, slog.String("document_file_id", m.Document.FileID))
	}
	if m.Photo != nil && len(m.Photo) > 0 {
		attrs = append(attrs, slog.String("photo_file_id", m.Photo[0].FileID))
	}
	if m.Video != nil {
		attrs = append(attrs, slog.String("video_file_id", m.Video.FileID))
	}
	if m.Audio != nil {
		attrs = append(attrs, slog.String("audio_file_id", m.Audio.FileID))
	}
	if m.Voice != nil {
		attrs = append(attrs, slog.String("voice_file_id", m.Voice.FileID))
	}

	return slog.GroupValue(attrs...)
}

// logChat returns a slog.Value for models.Chat for structured logging
func logChat(c *models.Chat) slog.Value {
	if c == nil {
		return slog.Value{}
	}

	attrs := []slog.Attr{
		slog.Int64("id", c.ID),
		slog.String("type", string(c.Type)),
	}

	if c.Title != "" {
		attrs = append(attrs, slog.String("title", c.Title))
	}
	if c.Username != "" {
		attrs = append(attrs, slog.String("username", c.Username))
	}
	if c.FirstName != "" {
		attrs = append(attrs, slog.String("first_name", c.FirstName))
	}
	if c.LastName != "" {
		attrs = append(attrs, slog.String("last_name", c.LastName))
	}
	if c.IsForum {
		attrs = append(attrs, slog.Bool("is_forum", c.IsForum))
	}
	if c.IsDirectMessages {
		attrs = append(attrs, slog.Bool("is_direct_messages", c.IsDirectMessages))
	}

	return slog.GroupValue(attrs...)
}

// LogUser returns a slog.Value for models.User for structured logging
func LogUser(u *models.User) slog.Value {
	if u == nil {
		return slog.Value{}
	}

	attrs := []slog.Attr{
		slog.Int64("id", u.ID),
		slog.Bool("is_bot", u.IsBot),
		slog.String("first_name", u.FirstName),
	}

	if u.LastName != "" {
		attrs = append(attrs, slog.String("last_name", u.LastName))
	}
	if u.Username != "" {
		attrs = append(attrs, slog.String("username", u.Username))
	}
	if u.LanguageCode != "" {
		attrs = append(attrs, slog.String("language_code", u.LanguageCode))
	}
	if u.IsPremium {
		attrs = append(attrs, slog.Bool("is_premium", u.IsPremium))
	}

	return slog.GroupValue(attrs...)
}
