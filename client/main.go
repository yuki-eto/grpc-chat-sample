package main

import (
	"context"
	"grpc-chat-sample/grpc"
	"log"
	"os"
	"os/signal"
	"syscall"

	googleGRPC "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := googleGRPC.Dial("localhost:19991", googleGRPC.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := grpc.NewChatServiceClient(conn)
	sigCh := make(chan os.Signal)

	ctx := context.TODO()
	var (
		token       string
		accessToken string
	)
	{
		req := &grpc.GetTokenRequest{Name: "test"}
		res, err := client.GetToken(ctx, req)
		if err != nil {
			panic(err)
		}
		log.Println(res.String())
		token = res.Token
		accessToken = res.AccessToken
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "login-token", token, "access-token", accessToken)

	{
		req := &grpc.JoinRoomRequest{
			RoomId: "room_test_id",
			Name:   "test_room",
		}
		res, err := client.JoinRoom(ctx, req)
		if err != nil {
			panic(err)
		}
		log.Printf("join: %s", res.String())
	}

	{
		req := &grpc.StreamRequest{
			RoomId: "room_test_id",
		}
		stream, err := client.Stream(ctx, req)
		if err != nil {
			panic(err)
		}

		go func() {
			for {
				res, err := stream.Recv()
				if err != nil {
					log.Printf("%+v", err)
					sigCh <- syscall.SIGTERM
					return
				}
				log.Printf("stream: %s", res.String())
			}
		}()
	}

	{
		req := &grpc.MessageRoomRequest{
			RoomId: "room_test_id",
			Text:   "hey!",
		}
		res, err := client.MessageRoom(ctx, req)
		if err != nil {
			panic(err)
		}
		log.Printf("chat: %s", res.String())
	}

	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGINT, os.Interrupt)
	<-sigCh

	{
		req := &grpc.LeaveRoomRequest{
			RoomId: "room_test_id",
		}
		res, err := client.LeaveRoom(ctx, req)
		if err != nil {
			log.Printf("leave error: %+v", err)
		}
		log.Printf("leave: %s", res.String())
	}

	log.Printf("done.")
}
