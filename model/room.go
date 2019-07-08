package model

import (
	"encoding/json"
	"grpc-chat-sample/dao"
	"grpc-chat-sample/entity"
	"grpc-chat-sample/grpc"
	"log"

	"github.com/cornelk/hashmap"
	"github.com/juju/errors"
)

type RoomInstance struct {
	*entity.Room

	players *PlayersInstance
	chats   *ChatsInstance

	joinCh  chan *PlayerInstance
	leaveCh chan *PlayerInstance
	chatCh  chan *ChatInstance
	doneCh  chan bool
}

func NewRoomInstance(room *entity.Room) *RoomInstance {
	r := &RoomInstance{
		Room:    room,
		players: NewPlayersInstance(),
		chats:   NewChatsInstance(),

		joinCh:  make(chan *PlayerInstance, 10),
		leaveCh: make(chan *PlayerInstance, 10),
		chatCh:  make(chan *ChatInstance, 10),
		doneCh:  make(chan bool, 1),
	}

	go r.handlingChannels()

	return r
}
func (r *RoomInstance) handlingChannels() {
	for {
		select {
		case <-r.doneCh:
			close(r.joinCh)
			close(r.leaveCh)
			close(r.chatCh)
			close(r.doneCh)
			log.Printf("closed")
			return
		case p, ok := <-r.joinCh:
			if !ok {
				continue
			}
			r.players.Set(p)
			r.players.AddStream(NewJoinStreamInstance(p))
			log.Printf("joined: %d, %v", p.ID, ok)
		case p, ok := <-r.leaveCh:
			if !ok {
				continue
			}
			r.players.Del(p.ID)
			r.players.AddStream(NewLeaveStreamInstance(p))
			log.Printf("leaved: %d", p.ID)
		case c, ok := <-r.chatCh:
			if !ok {
				continue
			}
			r.chats.Add(c)
			r.players.AddStream(NewChatStreamInstance(c))
			log.Printf("chat: %s", c.PlayerName)
		default:
		}
	}
}
func (r *RoomInstance) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(r.Room)
	return b, errors.Trace(err)
}
func (r *RoomInstance) Close() {
	r.players.Clear()
	r.chats.Clear()
	r.doneCh <- true
}
func (r *RoomInstance) Join(player *PlayerInstance) {
	go func() { r.joinCh <- player }()
}
func (r *RoomInstance) Leave(player *PlayerInstance) {
	go func() { r.leaveCh <- player }()
}
func (r *RoomInstance) Chat(chat *ChatInstance) {
	go func() { r.chatCh <- chat }()
}
func (r *RoomInstance) IsRoomPlayer(player *PlayerInstance) bool {
	return r.players.IsExists(player.ID)
}

func (r *RoomInstance) RoomForGRPC() *grpc.Room {
	return &grpc.Room{
		Id:   r.ID,
		Name: r.Name,
	}
}
func (r *RoomInstance) PlayersForGRPC() []*grpc.Player {
	return r.players.GRPCList()
}
func (r *RoomInstance) ChatsForGRPC() []*grpc.Chat {
	return r.chats.GRPCList()
}

type RoomRepository interface {
	Store(*RoomInstance) error
	Get(string) (*RoomInstance, error)
	Del(string)
}

type RoomRepositoryImpl struct {
	roomDao   dao.RoomDao
	instances *hashmap.HashMap
}

func NewRoomRepository() RoomRepository {
	return &RoomRepositoryImpl{
		roomDao:   dao.NewRoomDao(),
		instances: &hashmap.HashMap{},
	}
}

func (r *RoomRepositoryImpl) Store(room *RoomInstance) error {
	if err := r.roomDao.Insert(room.Room); err != nil {
		return errors.Trace(err)
	}
	r.instances.Set(room.ID, room)
	return nil
}

func (r *RoomRepositoryImpl) Get(id string) (*RoomInstance, error) {
	roomCache, ok := r.instances.Get(id)
	if ok {
		return roomCache.(*RoomInstance), nil
	}

	room, err := r.roomDao.Get(id)
	if err != nil {
		return nil, errors.Trace(err)
	}
	r.instances.Set(id, room)
	return NewRoomInstance(room), nil
}

func (r *RoomRepositoryImpl) Del(id string) {
	r.instances.Del(id)
	r.roomDao.Del(id)
}
