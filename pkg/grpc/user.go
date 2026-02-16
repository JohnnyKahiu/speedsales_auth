package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	pb "github.com/JohnnyKahiu/speed_sales_proto/user"
	"github.com/JohnnyKahiu/speedsales_login/pkg/users"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
}

func (s *UserServer) FetchUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	// run users validate token
	user, err := users.FetchUser(ctx, req.Username)
	if err != nil {
		if userStr, err := json.Marshal(user); err != nil {
			return &pb.UserResponse{
				UserDetails: string(userStr),
			}, nil
		}
	}

	fmt.Println("\n\t user = ", user)

	jstr, err := json.Marshal(user)
	if err != nil {
		return &pb.UserResponse{}, err
	}

	log.Println("grpc: AuthServer.FetchUser(): fetched user")
	return &pb.UserResponse{
		UserDetails: string(jstr),
	}, nil
}
