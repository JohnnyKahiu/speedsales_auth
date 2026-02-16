package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	pb "github.com/JohnnyKahiu/speed_sales_proto/user"
	"github.com/JohnnyKahiu/speedsales_login/pkg/users"

	"google.golang.org/grpc"
)

// AuthServer implements the AuthService gRPC interface
type AuthServer struct {
	pb.UnimplementedAuthServiceServer
}

// ValidateToken validates a token string
func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	// run users validate token
	details, auth := users.ValidateToken(ctx, req.Token)
	if !auth {
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	fmt.Println("\n\t details = ", details)

	fmt.Println("\n\n\t accept_payment =", details.AcceptPayment)
	fmt.Println("\t make_sales =", details.CashRollups)
	fmt.Println("\t till_num =", details.TillNum)

	jstr, err := json.Marshal(details)
	if err != nil {
		return &pb.ValidateTokenResponse{}, err
	}

	log.Println("grpc: AuthServer.ValidateToken(): validated token")
	return &pb.ValidateTokenResponse{
		Valid:  true,
		Rights: string(jstr),
	}, nil
}

func NewServer(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &AuthServer{})
	pb.RegisterUserServiceServer(grpcServer, &UserServer{})
	pb.RegisterTillServiceServer(grpcServer, &TillServer{})

	log.Println("Login service running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}
	return nil
}
