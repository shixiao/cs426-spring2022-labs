package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "cs426.yale.edu/lab1/video_rec_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	port            = flag.Int("port", 8080, "The server port")
	userServiceAddr = flag.String(
		"user-service",
		"[::1]:8081",
		"Server address for the UserService",
	)
	videoServiceAddr = flag.String(
		"video-service",
		"[::1]:8082",
		"Server address for the VideoService",
	)
	maxBatchSize = flag.Int(
		"batch-size",
		50,
		"Maximum size of batches sent to UserService and VideoService",
	)
)

type videoRecServiceServer struct {
	pb.UnimplementedVideoRecServiceServer
	// Add any data you want here
}

func makeVideoRecServiceServer() *videoRecServiceServer {
	return &videoRecServiceServer{}
}

func (server *videoRecServiceServer) GetTopVideos(
	ctx context.Context,
	req *pb.GetTopVideosRequest,
) (*pb.GetTopVideosResponse, error) {
	return nil, status.Error(
		codes.Unimplemented,
		"VideoRecService: unimplemented!",
	)
}

func main() {
	flag.Parse()
	log.Printf(
		"starting the server with flags: --user-service=%s --video-service=%s --batch-size=%d\n",
		*userServiceAddr,
		*videoServiceAddr,
		*maxBatchSize,
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	server := makeVideoRecServiceServer()
	pb.RegisterVideoRecServiceServer(s, server)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
