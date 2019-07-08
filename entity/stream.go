package entity

type StreamType string

const (
	StreamTypeChat  StreamType = "chat"
	StreamTypeJoin  StreamType = "join"
	StreamTypeLeave StreamType = "leave"
)

type Stream struct {
	Type   StreamType `json:"type"`
	Chat   *Chat      `json:"chat"`
	Player *Player    `json:"player"`
}
