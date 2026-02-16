package users

import (
	"context"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

type CustomClaims struct {
	UserDetails Users `json:"user_details"`
	jwt.RegisteredClaims
}

func (l *Users) ComparePassword(password string) (bool, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	/*
		compares hashed passwords and returns: (bool, bool, error)
			bool  := true or false if user is authenticated
			bool  := true or false if user is reset or not
			error := error thrown from running statement
	*/
	// search user
	user, err := FetchUser(ctx, l.Username)
	*l = user

	fmt.Println("\nhashed_pass =", l.password)
	fmt.Println("password =", password)
	fmt.Println("make_sales =", l.MakeSales)

	if err != nil {
		return false, false, err
	}
	if l.Reset {
		return true, true, nil
	} else if l.password == "" && l.Reset {
		return true, true, nil
	} else if l.password == "" && l.UserClass == "replication" {
		return true, false, nil
	}
	if l.password == "" {
		return false, false, nil
	}
	hash := l.password

	parts := strings.Split(hash, "$")

	c := &PasswordConfig{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &c.Memory, &c.Time, &c.Threads)
	if err != nil {
		return false, false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, false, err
	}
	c.KeyLen = uint32(len(decodedHash))

	comparisonHash := argon2.IDKey([]byte(password), salt, c.Time, c.Memory, c.Threads, c.KeyLen)

	return (subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1), false, nil
}

// GenerateJWT generates a JWT for the user
func (l *Users) GenerateJWT(privateKey *rsa.PrivateKey) (string, error) {
	log.Println(*l)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user_details": *l,
		"exp":          time.Now().Add(time.Hour * 24).Unix(),
		"iat":          time.Now().Unix(),
	})

	return token.SignedString(privateKey)
}

// ValidateJWT authenticates if JWT is valid and returns a true or false value
func (l *Users) ValidateJWT(tokenString string, publicKey *rsa.PublicKey) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure token is signed with RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	// Check if claim is expired
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return false
		}
	}

	// fmt.Println("claims =", claims)
	userBytes, _ := json.Marshal(claims["user_details"])
	// if err := json.Unmarshal(userBytes, &l); err != nil {
	// return false
	// }
	json.Unmarshal(userBytes, l)

	fmt.Println("\n\t user =", l.Username)
	fmt.Println("\t accept_payment =", l.AcceptPayment)
	fmt.Println("\t make_sales =", l.MakeSales)

	return true
}

// ValidateJWT authenticates if JWT is valid and returns a true or false value
func (l *Users) ValidateJWTs(tokenString string, publicKey *rsa.PublicKey) bool {
	claims := &CustomClaims{}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure token is signed with RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return false
	}

	log.Fatalln("claims = ", claims)

	if err != nil || !token.Valid {
		return false
	}

	*l = claims.UserDetails

	fmt.Println("\t user = ", l.Username)
	fmt.Println("\t accept_payment = ", l.AcceptPayment)
	fmt.Println("\t make_sales = ", l.MakeSales)

	return true
}

func (l *Users) Login(password string) (string, error) {
	// Generate JWT token
	privateKeyStr, err := os.ReadFile("private_key.pem")
	if err != nil {
		return "", err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyStr)
	if err != nil {
		return "", err
	}

	token, err := l.GenerateJWT(privateKey)
	if err != nil {
		return "", err
	}

	l.Token = token
	return token, nil
}

func ValidateToken(ctx context.Context, token string) (Users, bool) {
	publicKeyStr, err := os.ReadFile("public_key.pem")
	if err != nil {
		return Users{}, false
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyStr)
	if err != nil {
		log.Println("error, failed to parse public key with \t err =", err)
		return Users{}, false
	}

	fmt.Println("Login reset validateJWT")
	usr := Users{}
	authentic := usr.ValidateJWT(token, publicKey)
	if !authentic {
		return Users{}, false
	}

	return usr, true
}
