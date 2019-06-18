package main

import (
	"context"
	"flag"
	"net"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

var delay = flag.Duration("delay", 1*time.Millisecond, "time to sleep between starting goroutines")

func TestAPI(t *testing.T) {
	var s server
	ctx := context.Background()

	// Invoke API concurrently.
	var wg sync.WaitGroup
	wg.Add(2)

	hello := func(name string) {
		defer wg.Done()
		s.SayHello(ctx, &pb.HelloRequest{Name: name})
	}
	go hello("foo")
	time.Sleep(*delay)
	go hello("bar")
	wg.Wait()
}

func TestGRPC(t *testing.T) {
	var s server
	ctx := context.Background()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	g := grpc.NewServer()
	pb.RegisterGreeterServer(g, &s)
	go g.Serve(lis) // No error handling here, no waiting for server to stop.
	defer g.Stop()

	// Invoke via gRPC concurrently.
	var wg sync.WaitGroup
	wg.Add(2)

	conn, err := grpc.Dial("localhost"+port, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	hello := func(name string) {
		defer wg.Done()
		if _, err := c.SayHello(ctx, &pb.HelloRequest{Name: name}); err != nil {
			t.Fatalf("SayHello(%s): %v", name, err)
		}
	}
	go hello("foo")
	time.Sleep(*delay) // no data race reported for 100ms, reported for 1ms
	go hello("bar")
	wg.Wait()
}
