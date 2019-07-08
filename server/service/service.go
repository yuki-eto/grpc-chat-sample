package service

import (
	"context"
	"grpc-chat-sample/entity"
	"grpc-chat-sample/grpc"
	"grpc-chat-sample/model"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/juju/errors"
	"google.golang.org/grpc/metadata"
)

type ServerService struct {
	player model.PlayerRepository
	room   model.RoomRepository
}

func NewServerService() grpc.ChatServiceServer {
	return &ServerService{
		player: model.NewPlayerRepository(),
		room:   model.NewRoomRepository(),
	}
}

func (s *ServerService) GetToken(ctx context.Context, req *grpc.GetTokenRequest) (*grpc.GetTokenResult, error) {
	tokenUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Trace(err)
	}
	accessTokenUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Trace(err)
	}

	player := model.NewPlayerInstance(&entity.Player{
		Name:        req.Name,
		Token:       tokenUUID.String(),
		AccessToken: accessTokenUUID.String(),
	})
	if err := s.player.Store(player); err != nil {
		return nil, errors.Trace(err)
	}

	log.Printf("created user: %+v", player.Player)
	return &grpc.GetTokenResult{Token: player.Token, AccessToken: player.AccessToken}, nil
}

func (s *ServerService) getToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("cannot find token")
	}
	token := md.Get("login-token")
	if len(token) == 0 {
		return "", errors.New("cannot find token")
	}
	return token[0], nil
}
func (s *ServerService) getAccessToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("cannot find token")
	}
	token := md.Get("access-token")
	if len(token) == 0 {
		return "", errors.New("cannot find token")
	}
	return token[0], nil
}

func (s *ServerService) RefreshToken(ctx context.Context, req *grpc.RefreshTokenRequest) (*grpc.RefreshTokenResult, error) {
	token, err := s.getToken(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	player, err := s.player.Get(token)
	if err != nil {
		return nil, errors.Trace(err)
	}

	u, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Trace(err)
	}
	player.AccessToken = u.String()
	if err := s.player.Store(player); err != nil {
		return nil, errors.Trace(err)
	}

	return &grpc.RefreshTokenResult{AccessToken: player.AccessToken}, nil
}

func (s *ServerService) getPlayer(ctx context.Context) (*model.PlayerInstance, error) {
	token, err := s.getToken(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	player, err := s.player.Get(token)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if player.AccessToken != accessToken {
		return nil, errors.Unauthorizedf("access_token: %s", accessToken)
	}

	return player, nil
}

func (s *ServerService) JoinRoom(ctx context.Context, req *grpc.JoinRoomRequest) (*grpc.JoinRoomResult, error) {
	player, err := s.getPlayer(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	room, err := s.room.Get(req.RoomId)
	if err != nil && !errors.IsNotFound(err) {
		return nil, errors.Trace(err)
	}
	if errors.IsNotFound(err) {
		room = model.NewRoomInstance(&entity.Room{
			ID:   req.RoomId,
			Name: req.Name,
		})
		if err := s.room.Store(room); err != nil {
			return nil, errors.Trace(err)
		}
		log.Printf("created room: %+v", room.Room)
	}

	room.Join(player)
	res := &grpc.JoinRoomResult{
		Result:  true,
		Room:    room.RoomForGRPC(),
		Players: room.PlayersForGRPC(),
		Chats:   room.ChatsForGRPC(),
	}
	return res, nil
}

func (s *ServerService) LeaveRoom(ctx context.Context, req *grpc.LeaveRoomRequest) (*grpc.LeaveRoomResult, error) {
	player, err := s.getPlayer(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	room, err := s.room.Get(req.RoomId)
	if err != nil && !errors.IsNotFound(err) {
		return nil, errors.Trace(err)
	}
	if !room.IsRoomPlayer(player) {
		return nil, errors.BadRequestf("not in room: %s", req.RoomId)
	}

	room.Leave(player)
	return &grpc.LeaveRoomResult{Result: true}, nil
}

func (s *ServerService) MessageRoom(ctx context.Context, req *grpc.MessageRoomRequest) (*grpc.MessageRoomResult, error) {
	player, err := s.getPlayer(ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}

	room, err := s.room.Get(req.RoomId)
	if err != nil && !errors.IsNotFound(err) {
		return nil, errors.Trace(err)
	}

	if !room.IsRoomPlayer(player) {
		return nil, errors.BadRequestf("cannot send to different room: %s", req.RoomId)
	}

	now := time.Now()
	room.Chat(model.NewChatInstance(&entity.Chat{
		PlayerName: player.Name,
		Text:       req.Text,
		Time:       &now,
	}))

	return &grpc.MessageRoomResult{Result: true}, nil
}

func (s *ServerService) Stream(req *grpc.StreamRequest, stream grpc.ChatService_StreamServer) error {
	ctx := stream.Context()
	player, err := s.getPlayer(ctx)
	if err != nil {
		return errors.Trace(err)
	}

	room, err := s.room.Get(req.RoomId)
	if err != nil {
		return errors.Trace(err)
	}
	if !room.IsRoomPlayer(player) {
		return errors.BadRequestf("cannot send to different room: %s", req.RoomId)
	}

	for {
		st, ok := player.PopStream()
		if !ok {
			return nil
		}

		log.Printf("stream: %+v", st.Stream)
		res := s.handlingStream(st)
		if err := stream.Send(res); err != nil {
			return errors.Trace(err)
		}
	}
}

func (s *ServerService) handlingStream(stream *model.StreamInstance) *grpc.StreamResponse {
	var res *grpc.StreamResponse
	switch stream.Type {
	case entity.StreamTypeChat:
		res = &grpc.StreamResponse{
			Type: grpc.StreamResponse_Chat,
			Chat: &grpc.Chat{
				PlayerName: stream.Chat.PlayerName,
				Text:       stream.Chat.Text,
				Time:       stream.Chat.Time.Unix(),
			},
		}
	case entity.StreamTypeJoin, entity.StreamTypeLeave:
		res = &grpc.StreamResponse{
			Player: &grpc.Player{
				Id:   stream.Player.ID,
				Name: stream.Player.Name,
			},
		}
		if stream.Type == entity.StreamTypeJoin {
			res.Type = grpc.StreamResponse_Joined
		} else {
			res.Type = grpc.StreamResponse_Leaved
		}
	}
	return res
}
