package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/JohnnyKahiu/speedsales_login/pkg/license"
)

// LicensePOST handles POST /license
// Body: {"company_id": 1}  → issues a new 6-month license
// Body: {"company_id": 1, "action": "revoke"} → revokes all licenses for company
func LicensePOST(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	if r.Method == "OPTIONS" {
		return
	}

	respMap := make(map[string]interface{})

	b, err := io.ReadAll(r.Body)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "could not read request body"
		json.NewEncoder(w).Encode(respMap)
		return
	}

	var args map[string]string
	if err = json.Unmarshal(b, &args); err != nil {
		respMap["response"] = "error"
		respMap["message"] = "invalid JSON"
		json.NewEncoder(w).Encode(respMap)
		return
	}

	companyID, err := strconv.ParseInt(args["company_id"], 10, 64)
	if err != nil || companyID == 0 {
		respMap["response"] = "error"
		respMap["message"] = "company_id required"
		json.NewEncoder(w).Encode(respMap)
		return
	}

	if args["action"] == "revoke" {
		if err = license.Revoke(companyID); err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to revoke license"
			json.NewEncoder(w).Encode(respMap)
			return
		}
		respMap["response"] = "success"
		respMap["message"] = "license revoked"
		json.NewEncoder(w).Encode(respMap)
		return
	}

	l, err := license.New(companyID)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "failed to issue license"
		json.NewEncoder(w).Encode(respMap)
		return
	}

	respMap["response"] = "success"
	respMap["license"] = l
	json.NewEncoder(w).Encode(respMap)
}

// LicenseGET handles GET /license?company_id=1
// Returns the active license and its validity status for the given company.
func LicenseGET(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)
	if r.Method == "OPTIONS" {
		return
	}

	respMap := make(map[string]interface{})

	companyID, err := strconv.ParseInt(r.URL.Query().Get("company_id"), 10, 64)
	if err != nil || companyID == 0 {
		respMap["response"] = "error"
		respMap["message"] = "company_id query param required"
		json.NewEncoder(w).Encode(respMap)
		return
	}

	l, err := license.FetchByCompany(companyID)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "no active license found"
		json.NewEncoder(w).Encode(respMap)
		return
	}

	_, expired, _ := license.Validate(companyID)

	respMap["response"] = "success"
	respMap["license"] = l
	respMap["expired"] = expired
	json.NewEncoder(w).Encode(respMap)
}
