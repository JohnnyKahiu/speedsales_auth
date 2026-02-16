package proto

import (
	"context"
	"encoding/json"
	"log"

	authpb "github.com/JohnnyKahiu/speed_sales_proto/auth"
	"github.com/JohnnyKahiu/speedsales_login/pkg/users"
)

// AuthServer implements auth.v1.AuthService
type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
}

// ValidateToken validates authentication tokens
func (s *AuthServer) ValidateToken(
	ctx context.Context,
	req *authpb.ValidateTokenRequest,
) (*authpb.ValidateTokenResponse, error) {

	user, isValid := users.ValidateToken(ctx, req.Token)

	log.Println("Login service: token validation request")

	jsonUser, _ := json.Marshal(user)

	// Replace with JWT / OAuth logic
	if isValid {
		return &authpb.ValidateTokenResponse{
			Valid:    true,
			Username: user.Username,
			Rights:   string(jsonUser),
		}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid: false,
	}, nil
}
