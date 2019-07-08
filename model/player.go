package model

import (
	"grpc-chat-sample/dao"
	"grpc-chat-sample/entity"
	"grpc-chat-sample/grpc"

	"github.com/cornelk/hashmap"
	"github.com/juju/errors"
)

type PlayerInstance struct {
	*entity.Player

	streamCh chan *StreamInstance
}

func NewPlayerInstance(player *entity.Player) *PlayerInstance {
	return &PlayerInstance{
		Player: player,

		streamCh: make(chan *StreamInstance, 100),
	}
}
func (p *PlayerInstance) Clear() {
	close(p.streamCh)
	p.streamCh = make(chan *StreamInstance, 100)
}
func (p *PlayerInstance) AddStream(stream *StreamInstance) {
	go func() { p.streamCh <- stream }()
}
func (p *PlayerInstance) PopStream() (*StreamInstance, bool) {
	if p.streamCh == nil {
		return nil, false
	}
	instance, ok := <-p.streamCh
	return instance, ok
}

type PlayersInstance struct {
	values *hashmap.HashMap
}

func NewPlayersInstance() *PlayersInstance {
	players := new(PlayersInstance)
	players.Clear()
	return players
}
func (p *PlayersInstance) Clear() {
	if p.values != nil {
		p.Each(func(player *PlayerInstance) { player.Clear() })
	}
	p.values = &hashmap.HashMap{}
}
func (p *PlayersInstance) Set(player *PlayerInstance) {
	p.values.Set(player.ID, player)
}
func (p *PlayersInstance) IsExists(id uint64) bool {
	_, exists := p.values.Get(id)
	return exists
}
func (p *PlayersInstance) Get(id uint64) (*PlayerInstance, bool) {
	player, ok := p.values.Get(id)
	if !ok {
		return nil, false
	}
	return player.(*PlayerInstance), true
}
func (p *PlayersInstance) Del(id uint64) {
	player, exists := p.Get(id)
	if !exists {
		return
	}
	p.values.Del(id)
	player.Clear()
}
func (p *PlayersInstance) Each(f func(player *PlayerInstance)) {
	for kv := range p.values.Iter() {
		f(kv.Value.(*PlayerInstance))
	}
}
func (p *PlayersInstance) EachWithError(f func(player *PlayerInstance) error) error {
	for kv := range p.values.Iter() {
		if err := f(kv.Value.(*PlayerInstance)); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}
func (p *PlayersInstance) IsEmpty() bool {
	return p.values.Len() == 0
}
func (p *PlayersInstance) AddStream(stream *StreamInstance) {
	p.Each(func(player *PlayerInstance) {
		player.AddStream(stream)
	})
}
func (p *PlayersInstance) GRPCList() (list []*grpc.Player) {
	p.Each(func(player *PlayerInstance) {
		list = append(list, &grpc.Player{
			Id:   player.ID,
			Name: player.Name,
		})
	})
	return list
}

type PlayerRepository interface {
	Store(*PlayerInstance) error
	Get(string) (*PlayerInstance, error)
}
type PlayerRepositoryImpl struct {
	playerDao dao.PlayerDao
	instances *hashmap.HashMap
}

func NewPlayerRepository() PlayerRepository {
	return &PlayerRepositoryImpl{
		playerDao: dao.NewPlayerDao(),
		instances: &hashmap.HashMap{},
	}
}

func (p *PlayerRepositoryImpl) Store(player *PlayerInstance) error {
	if player.ID == 0 {
		if err := p.playerDao.Insert(player.Player); err != nil {
			return errors.Trace(err)
		}
	} else {
		if err := p.playerDao.Update(player.Player); err != nil {
			return errors.Trace(err)
		}
	}
	p.instances.Set(player.Token, player)
	return nil
}

func (p *PlayerRepositoryImpl) Get(token string) (*PlayerInstance, error) {
	playerCache, ok := p.instances.Get(token)
	if ok {
		return playerCache.(*PlayerInstance), nil
	}

	player, err := p.playerDao.GetByToken(token)
	if err != nil {
		return nil, errors.Trace(err)
	}
	p.instances.Set(player.Token, player)
	return NewPlayerInstance(player), nil
}
