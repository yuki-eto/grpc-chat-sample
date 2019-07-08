package dao

import (
	"grpc-chat-sample/entity"

	"github.com/cornelk/hashmap"
	"github.com/juju/errors"
)

type PlayerDao interface {
	Insert(*entity.Player) error
	Update(*entity.Player) error
	GetByToken(string) (*entity.Player, error)
}

func NewPlayerDao() PlayerDao {
	return playerDao
}

type PlayerDaoImpl struct {
	players       *hashmap.HashMap
	playerByToken *hashmap.HashMap

	incrementID *IncrementID
}

func (p *PlayerDaoImpl) Insert(player *entity.Player) error {
	if _, exists := p.playerByToken.Get(player.Token); exists {
		return errors.AlreadyExistsf("token: %s", player.Token)
	}
	player.ID = p.incrementID.Get()
	p.players.Set(player.ID, player)
	p.playerByToken.Set(player.Token, player)
	return nil
}

func (p *PlayerDaoImpl) Update(player *entity.Player) error {
	p.players.Set(player.ID, player)
	p.playerByToken.Set(player.Token, player)
	return nil
}

func (p *PlayerDaoImpl) GetByToken(token string) (*entity.Player, error) {
	player, ok := p.playerByToken.Get(token)
	if !ok {
		return nil, errors.NotFoundf("token: %s", token)
	}
	return player.(*entity.Player), nil
}

var playerDao PlayerDao

func init() {
	playerDao = &PlayerDaoImpl{
		players:       &hashmap.HashMap{},
		playerByToken: &hashmap.HashMap{},
		incrementID:   NewIncrementID(),
	}
}
