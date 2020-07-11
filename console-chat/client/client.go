package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"wesionary.team/dipeshdulal/console-chat/chatpb"
)

var channelName = flag.String("channel", "default", "Channel name for chatting")
var senderName = flag.String("sender", "default", "Senders name")
var tcpServer = flag.String("server", ":5400", "Tcp server")

func joinChannel(ctx context.Context, client chatpb.ChatServiceClient) {

	channel := chatpb.Channel{Name: *channelName, SendersName: *senderName}
	stream, err := client.JoinChannel(ctx, &channel)
	if err != nil {
		log.Fatalf("client.JoinChannel(ctx, &channel) throws: %v", err)
	}

	fmt.Printf("Joined channel: %v \n", *channelName)

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive message from channel joining. \nErr: %v", err)
			}

			if *senderName != in.Sender {
				fmt.Printf("MESSAGE: (%v) -> %v \n", in.Sender, in.Message)
			}
		}
	}()

	<-waitc
}

func sendMessage(ctx context.Context, client chatpb.ChatServiceClient, message string) {
	stream, err := client.SendMessage(ctx)
	if err != nil {
		log.Printf("Cannot send message: error: %v", err)
	}
	msg := chatpb.Message{
		Channel: &chatpb.Channel{
			Name:        *channelName,
			SendersName: *senderName},
		Message: message,
		Sender:  *senderName,
	}
	stream.Send(&msg)

	ack, err := stream.CloseAndRecv()
	fmt.Printf("Message sent: %v \n", ack)
}

func main() {

	flag.Parse()

	fmt.Println("--- CLIENT APP ---")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())

	conn, err := grpc.Dial(*tcpServer, opts...)
	if err != nil {
		log.Fatalf("Fail to dail: %v", err)
	}

	defer conn.Close()

	ctx := context.Background()
	client := chatpb.NewChatServiceClient(conn)

	go joinChannel(ctx, client)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		go sendMessage(ctx, client, scanner.Text())
	}

}
