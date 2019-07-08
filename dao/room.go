package dao

import (
	"grpc-chat-sample/entity"

	"github.com/cornelk/hashmap"
	"github.com/juju/errors"
)

type RoomDao interface {
	Insert(*entity.Room) error
	Update(*entity.Room) error
	Get(string) (*entity.Room, error)
	Del(string)
}

type RoomDaoImpl struct {
	rooms *hashmap.HashMap
}

func NewRoomDao() RoomDao {
	return roomDao
}

func (r *RoomDaoImpl) Insert(room *entity.Room) error {
	if !r.rooms.Insert(room.ID, room) {
		return errors.AlreadyExistsf("room_id: %s", room.ID)
	}
	return nil
}

func (r *RoomDaoImpl) Update(room *entity.Room) error {
	r.rooms.Set(room.ID, room)
	return nil
}

func (r *RoomDaoImpl) Get(id string) (*entity.Room, error) {
	room, ok := r.rooms.Get(id)
	if !ok {
		return nil, errors.NotFoundf("room_id: %s", id)
	}
	return room.(*entity.Room), nil
}

func (r *RoomDaoImpl) Del(id string) {
	r.rooms.Del(id)
}

var roomDao RoomDao

func init() {
	roomDao = &RoomDaoImpl{
		rooms: &hashmap.HashMap{},
	}
}
