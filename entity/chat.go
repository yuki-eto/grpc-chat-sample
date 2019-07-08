package entity

import (
	"time"
)

type Chat struct {
	PlayerName string     `json:"player_name"`
	Text       string     `json:"text"`
	Time       *time.Time `json:"time"`
}
