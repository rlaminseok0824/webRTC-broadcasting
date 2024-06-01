package grpc_server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/webRTC-broadcasting/handler"
	pb "github.com/webRTC-broadcasting/proto"
)

type server struct {
    pb.UnimplementedGetTrackLocalServiceServer
    // trackLocals map[string]*webrtc.TrackLocalStaticRTP
    // listLock sync.RWMutex
}

func StartGRPCServer(){
	lis,err := net.Listen("tcp", ":4040")
	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer()
	pb.RegisterGetTrackLocalServiceServer(srv, &server{})
	reflection.Register(srv)


	if e := srv.Serve(lis); e != nil {
		panic(err)
	}
}

func (s *server) GetTrackLocal(ctx context.Context, req *pb.Request) (*pb.TrackLocalResponse, error) {
    handler.ListLock.RLock()
    defer handler.ListLock.RUnlock()

	fmt.Println(handler.TrackLocals)

    var trackIds []string
    for id := range handler.TrackLocals {
        trackIds = append(trackIds, id)
    }

    return &pb.TrackLocalResponse{TrackIds: trackIds}, nil
}