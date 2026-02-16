package users

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateJWT_Reproduction(t *testing.T) {
	// 1. Generate RSA Keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// 2. Create a User with data
	originalUser := Users{
		Username:      "testuser",
		FirstName:     "Test",
		LastName:      "User",
		MakeSales:     true,
		AcceptPayment: true,
		CompanyID:     100,
	}

	// 3. Generate Token using the method in authenticate.go
	// Note: We need to mock 'l' as originalUser
	tokenStr, err := originalUser.GenerateJWT(privateKey)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}
	fmt.Printf("Generated Token: %s\n", tokenStr)

	// 4. Validate Token
	// We use a NEW empty user struct as receiver, like in the usage code
	validatingUser := &Users{}

	// Print validatingUser before
	fmt.Printf("Before ValidateJWT: %+v\n", validatingUser)

	valid := validatingUser.ValidateJWT(tokenStr, publicKey)

	if !valid {
		t.Errorf("ValidateJWT returned false, expected true")
	}

	// Print validatingUser after
	fmt.Printf("After ValidateJWT: %+v\n", validatingUser)

	// 5. Assertions
	if validatingUser.Username != originalUser.Username {
		t.Errorf("Username mismatch. Expected %s, got %s", originalUser.Username, validatingUser.Username)
	}
	if validatingUser.MakeSales != originalUser.MakeSales {
		t.Errorf("MakeSales mismatch. Expected %v, got %v", originalUser.MakeSales, validatingUser.MakeSales)
	}
}

// Manual verify steps to debug internals if needed
func TestManualMarshaling(t *testing.T) {
	originalUser := Users{
		Username:  "testuser",
		MakeSales: true,
	}

	// Emulate what jwt does (roughly)
	claims := jwt.MapClaims{
		"user_details": originalUser,
		"exp":          time.Now().Add(time.Hour).Unix(),
	}

	// Create token to see how it marshals
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signedStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	// Parse it back
	parsedToken, _ := jwt.Parse(signedStr, nil) // ignore sig check for this test
	parsedClaims := parsedToken.Claims.(jwt.MapClaims)

	userDetails := parsedClaims["user_details"]
	fmt.Printf("Type of user_details in claims: %T\n", userDetails)
	fmt.Printf("Value of user_details: %+v\n", userDetails)

	// Emulate ValidateJWT logic
	user := Users{}
	userBytes, err := json.Marshal(userDetails)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	fmt.Printf("JSON Bytes: %s\n", string(userBytes))

	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if user.Username != originalUser.Username {
		t.Errorf("Username mismatch after round trip")
	}
}
