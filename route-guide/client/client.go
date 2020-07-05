package main

import (
	"context"
	"google.golang.org/grpc"
	"io"
	"log"
	"math/rand"
	"time"
	"wesionary.team/dipeshdulal/route-guide/mrouteguide"
)

func printFeature(client mrouteguide.RouteGuideClient, point *mrouteguide.Point) {
	log.Printf("Getting feature for point (%d, %d)", point.Latitude, point.Longitude)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	feature, err := client.GetFeature(ctx, point)
	if err != nil {
		log.Fatalf("%v.GetFeatures(_) = _, %v ", client, err)
	}

	log.Println(feature)
}

// printFeatures list all the features within the given bounding rectangle
func printFeatures(client mrouteguide.RouteGuideClient, rect *mrouteguide.Rectangle) {
	log.Printf("Looking for feature within %v", rect)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.ListFeatures(ctx, rect)
	if err != nil {
		log.Fatalf("%v.ListFeatures(_) = _, %v ", client, err)
	}
	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}
		log.Println(feature)
	}
}

func runRecordRoute(client mrouteguide.RouteGuideClient) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	pointCount := int(r.Int31n(100)) + 2
	var points []*mrouteguide.Point
	for i := 0; i < pointCount; i++ {
		points = append(points, randomPoint(r))
	}

	log.Printf("Traversing %d points. ", len(points))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.RecordRoute(ctx)
	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v ", client, err)
	}
	for _, point := range points {
		if err := stream.Send(point); err != nil {
			log.Fatalf("%v.Send(%v) = %v", stream, point, err)
		}
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}

	log.Printf("Route summary: %v", reply)
}

// runRouteChat receives a sequence of route notes, while sending notes for various locations.
func runRouteChat(client mrouteguide.RouteGuideClient) {
	notes := []*mrouteguide.RouteNote{
		{Location: &mrouteguide.Point{Latitude: 0, Longitude: 1}, Message: "First message"},
		{Location: &mrouteguide.Point{Latitude: 0, Longitude: 2}, Message: "Second message"},
		{Location: &mrouteguide.Point{Latitude: 0, Longitude: 3}, Message: "Third message"},
		{Location: &mrouteguide.Point{Latitude: 0, Longitude: 1}, Message: "Fourth message"},
		{Location: &mrouteguide.Point{Latitude: 0, Longitude: 2}, Message: "Fifth message"},
		{Location: &mrouteguide.Point{Latitude: 0, Longitude: 3}, Message: "Sixth message"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.RouteChat(ctx)
	if err != nil {
		log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("Got message %s at point(%d, %d)", in.Message, in.Location.Latitude, in.Location.Longitude)
		}
	}()
	for _, note := range notes {
		if err := stream.Send(note); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	}
	stream.CloseSend()
	<-waitc
}

func randomPoint(r *rand.Rand) *mrouteguide.Point {
	lat := (r.Int31n(180) - 90) * 1e7
	long := (r.Int31n(360) - 180) * 1e7
	return &mrouteguide.Point{Latitude: lat, Longitude: long}
}

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
	conn, err := grpc.Dial(":3000", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := mrouteguide.NewRouteGuideClient(conn)

	// Looking for a valid feature
	printFeature(client, &mrouteguide.Point{Latitude: 409146138, Longitude: -746188906})

	// Feature missing.
	printFeature(client, &mrouteguide.Point{Latitude: 0, Longitude: 0})

	// Looking for features between 40, -75 and 42, -73.
	printFeatures(client, &mrouteguide.Rectangle{
		Lo: &mrouteguide.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &mrouteguide.Point{Latitude: 420000000, Longitude: -730000000},
	})

	// RecordRoute
	runRecordRoute(client)

	// RouteChat
	runRouteChat(client)
}
