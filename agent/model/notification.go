package model

type Notification struct {
	ChannelLevel string `json:"channel_level"`
	ChannelName  string `json:"channel_name"`
	AlertOn      string `json:"alert_on"`
}
