package authentication

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/JohnnyKahiu/speedsales_login/pkg/license"
	"github.com/JohnnyKahiu/speedsales_login/pkg/users"
	"github.com/golang-jwt/jwt/v5"
)

// contextKey is an unexported type so keys set here can never collide with a
// context value set by another package using a plain string key.
type contextKey string

// ParamsContextKey is where RateLimiterMiddleware stashes the already-
// parsed /login request body — the middleware has to fully drain r.Body to
// read the username for rate-limiting before POST ever sees the request, so
// this is the handoff contract between the two.
const ParamsContextKey contextKey = "params"

func POST(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	// get post body items
	args, ok := r.Context().Value(ParamsContextKey).(map[string]string)
	if !ok {
		respMap["response"] = "error"
		respMap["message"] = "params error"
		return respMap
	}

	// Unmarshal into an args map

	user := users.Users{Username: args["username"]}

	// compare argon2 harshed password
	match, reset, err := user.ComparePassword(args["password"])
	if !match || err != nil {
		respMap["response"] = "error" // return success if no error and match
		respMap["message"] = "wrong username or password"
		return respMap
	}

	// A reset-required account authenticates without proving its real password,
	// so it must not carry any rights until it's gone through /login/reset.
	if reset {
		stripAllRights(&user)
	}

	fmt.Println("username =", user.FirstName, " last_name =", user.LastName)
	fmt.Println("accept_payment =", user.AcceptPayment)
	fmt.Println("make_sales =", user.MakeSales)
	fmt.Println("till_num =", user.TillNum)
	fmt.Println("audit_stock =", user.AuditStock)

	privateKeyStr, err := os.ReadFile("private_key.pem")
	if err != nil {
		log.Println("error reading private key =", err)
		respMap["response"] = "error"
		respMap["message"] = "fatal error"
		return respMap
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyStr)
	if err != nil {
		log.Println("error parsing private key    err =", err)
		respMap["response"] = "error"
		respMap["message"] = "could not parse private key"
		return respMap
	}

	// check license validity for this company
	licenseExpired := false
	_, expired, licErr := license.Validate(user.CompanyID)
	if licErr != nil {
		log.Println("warning: license check failed    err =", licErr)
	}
	if expired {
		licenseExpired = true
		stripToBasicRights(&user)
	}

	// generate jwt token
	token, _ := user.GenerateJWT(privateKey)
	if reset {
		respMap["response"] = "reset"
		respMap["token"] = fmt.Sprintf("%v", token)
		respMap["username"] = user.Username
		respMap["message"] = "reset user password"
		respMap["user_details"] = user
		return respMap
	}

	respMap["response"] = "success"
	respMap["token"] = fmt.Sprintf("%v", token)
	respMap["username"] = user.Username
	respMap["till_num"] = user.TillNum
	respMap["user_details"] = user
	respMap["license_expired"] = licenseExpired
	return respMap
}

// stripAllRights reduces u to bare identity — no permissions at all. Used for
// reset-required logins, which haven't proven the account's real password yet.
func stripAllRights(u *users.Users) {
	*u = users.Users{
		AutoId:    u.AutoId,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
		Branch:    u.Branch,
		CompanyID: u.CompanyID,
		UserClass: u.UserClass,
		Profile:   u.Profile,
		Reset:     u.Reset,
	}
}

// stripToBasicRights zeroes all permissions on u except make_sales and accept_payment.
func stripToBasicRights(u *users.Users) {
	makeSales := u.MakeSales
	acceptPayment := u.AcceptPayment

	*u = users.Users{
		AutoId:        u.AutoId,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Username:      u.Username,
		Branch:        u.Branch,
		CompanyID:     u.CompanyID,
		UserClass:     u.UserClass,
		TillNum:       u.TillNum,
		StkLocation:   u.StkLocation,
		Profile:       u.Profile,
		MakeSales:     makeSales,
		AcceptPayment: acceptPayment,
	}
}
