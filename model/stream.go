package model

import (
	"grpc-chat-sample/entity"
)

type StreamInstance struct {
	*entity.Stream
}

func NewJoinStreamInstance(player *PlayerInstance) *StreamInstance {
	return &StreamInstance{
		Stream: &entity.Stream{
			Type:   entity.StreamTypeJoin,
			Player: player.Player,
		},
	}
}
func NewLeaveStreamInstance(player *PlayerInstance) *StreamInstance {
	return &StreamInstance{
		Stream: &entity.Stream{
			Type:   entity.StreamTypeLeave,
			Player: player.Player,
		},
	}
}
func NewChatStreamInstance(chat *ChatInstance) *StreamInstance {
	return &StreamInstance{
		Stream: &entity.Stream{
			Type: entity.StreamTypeChat,
			Chat: chat.Chat,
		},
	}
}
