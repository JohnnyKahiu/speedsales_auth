package grpc

import (
	"context"
	"fmt"
	"log"

	pb "github.com/JohnnyKahiu/speed_sales_proto/user"
	"github.com/JohnnyKahiu/speedsales_login/pkg/users"
)

// type TillServer struct {
// 	pb.UnimplementedTillServiceServer
// }

// func (s *TillServer) FetchTill(ctx context.Context, req *pb.TillRequest) (*pb.TillResponse, error) {
// 	return &pb.TillResponse{}, nil
// }

type TillServer struct {
	pb.UnimplementedTillServiceServer
}

func (s *TillServer) UpdateTill(ctx context.Context, req *pb.UpdateTillRequest) (*pb.UpdateTillResponse, error) {
	// run users validate token
	err := users.UpdateTill(ctx, req.Username, fmt.Sprintf("%v", req.TillNum))
	if err != nil {
		return &pb.UpdateTillResponse{
			Success:  false,
			Message:  "failed to update till",
			Response: "failed",
		}, err
	}

	log.Println("grpc: AuthServer.FetchUser(): fetched user")
	return &pb.UpdateTillResponse{
		Success:  true,
		Message:  "till updated to user successfully",
		Response: "success",
	}, nil
}
