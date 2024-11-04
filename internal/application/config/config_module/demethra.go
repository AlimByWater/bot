package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type Demethra struct {
	BotToken                     string
	BotName                      string
	ChatIDForLogs                int64
	ElysiumFmID                  int64
	ElysiumForumID               int64
	ElysiumFmCommentID           int64
	TracksDbChannel              int64
	CurrentTrackMessageID        int
	ListenerIdleTimeoutInMinutes int
	TelegramBotApiServer         string
}

func NewDemethraConfig() *Demethra {
	return &Demethra{}
}

func (c Demethra) GetBotToken() string                  { return c.BotToken }
func (c Demethra) GetBotName() string                   { return c.BotName }
func (c Demethra) GetChatIDForLogs() int64              { return c.ChatIDForLogs }
func (c Demethra) GetElysiumFmID() int64                { return c.ElysiumFmID }
func (c Demethra) GetElysiumForumID() int64             { return c.ElysiumForumID }
func (c Demethra) GetElysiumFmCommentID() int64         { return c.ElysiumFmCommentID }
func (c Demethra) GetTracksDbChannel() int64            { return c.TracksDbChannel }
func (c Demethra) GetCurrentTrackMessageID() int        { return c.CurrentTrackMessageID }
func (c Demethra) GetListenerIdleTimeoutInMinutes() int { return c.ListenerIdleTimeoutInMinutes }
func (c Demethra) GetTelegramBotApiServer() string      { return c.TelegramBotApiServer }

func (c Demethra) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.BotToken, validation.Required),
		validation.Field(&c.BotName, validation.Required),
		validation.Field(&c.ChatIDForLogs, validation.Required),
		validation.Field(&c.ElysiumFmID, validation.Required),
		validation.Field(&c.ElysiumForumID, validation.Required),
		validation.Field(&c.ElysiumFmCommentID, validation.Required),
		validation.Field(&c.TracksDbChannel, validation.Required),
		validation.Field(&c.CurrentTrackMessageID, validation.Required),
		validation.Field(&c.ListenerIdleTimeoutInMinutes, validation.Required),
	)
}
