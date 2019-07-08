package main

import (
	"fmt"
	"grpc-chat-sample/grpc"
	"grpc-chat-sample/server/service"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	gGrpc "google.golang.org/grpc"
)

func main() {
	port := 19991
	addr := fmt.Sprintf("localhost:%d", port)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	grpcServer := gGrpc.NewServer()
	grpc.RegisterChatServiceServer(grpcServer, service.NewServerService())
	log.Printf("listen gRPC on %s", addr)

	go func() {
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatalf("error: %+v", err)
		}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGINT, os.Interrupt)
	<-sigCh

	log.Printf("shutdown")
	grpcServer.Stop()
}
