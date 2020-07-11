package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"wesionary.team/dipeshdulal/console-chat/chatpb"
)

type chatServiceServer struct {
	chatpb.UnimplementedChatServiceServer
	mu      sync.Mutex
	channel map[string][]chan *chatpb.Message
}

func (s *chatServiceServer) JoinChannel(ch *chatpb.Channel, msgStream chatpb.ChatService_JoinChannelServer) error {

	msgChannel := make(chan *chatpb.Message)
	s.channel[ch.Name] = append(s.channel[ch.Name], msgChannel)

	// doing this never closes the stream
	for {
		select {
		case <-msgStream.Context().Done():
			return nil
		case msg := <-msgChannel:
			fmt.Printf("GO ROUTINE (got message): %v \n", msg)
			msgStream.Send(msg)
		}
	}
}

func (s *chatServiceServer) SendMessage(msgStream chatpb.ChatService_SendMessageServer) error {
	msg, err := msgStream.Recv()

	if err == io.EOF {
		return nil
	}

	if err != nil {
		return err
	}

	fmt.Printf("Received Msg: %v \n", msg)

	ack := chatpb.MessageAck{Status: "SENT"}
	msgStream.SendAndClose(&ack)

	go func() {
		streams := s.channel[msg.Channel.Name]
		for idx, msgChan := range streams {
			fmt.Println("idx: ", idx)
			msgChan <- msg
			fmt.Println("After here")
		}
	}()

	return nil
}

func newServer() *chatServiceServer {
	s := &chatServiceServer{
		channel: make(map[string][]chan *chatpb.Message),
	}
	fmt.Println(s)
	return s
}

func main() {
	fmt.Println("--- SERVER APP ---")
	lis, err := net.Listen("tcp", "localhost:5400")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	chatpb.RegisterChatServiceServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
