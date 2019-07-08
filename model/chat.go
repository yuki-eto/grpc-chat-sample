package model

import (
	"grpc-chat-sample/dao"
	"grpc-chat-sample/entity"
	"grpc-chat-sample/grpc"
)

type ChatInstance struct {
	*entity.Chat
}

func NewChatInstance(chat *entity.Chat) *ChatInstance {
	return &ChatInstance{
		Chat: chat,
	}
}

type ChatsInstance struct {
	linkedList *dao.LinkedList
}

func NewChatsInstance() *ChatsInstance {
	return &ChatsInstance{
		linkedList: dao.NewLinkedList(),
	}
}
func (c *ChatsInstance) Clear() {
	c.linkedList = &dao.LinkedList{}
}
func (c ChatsInstance) Add(chat *ChatInstance) {
	c.linkedList.Add(chat)
}
func (c ChatsInstance) GRPCList() (list []*grpc.Chat) {
	c.linkedList.Each(func(val *dao.LinkedListValue) {
		chat := val.Value.(*ChatInstance)
		list = append(list, &grpc.Chat{
			PlayerName: chat.PlayerName,
			Text:       chat.Text,
			Time:       chat.Time.Unix(),
		})
	})
	return list
}
