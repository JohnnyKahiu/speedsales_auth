package license

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_login/pkg/database"
)

// License represents a software license tied to a company.
type License struct {
	AutoID     int64     `json:"auto_id"`
	LicenseKey string    `json:"license_key"`
	CompanyID  int64     `json:"company_id"`
	IssuedAt   time.Time `json:"issued_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsActive   bool      `json:"is_active"`
}

// GenTable creates the licenses table if it does not exist.
func GenTable() error {
	sql := `
		CREATE TABLE IF NOT EXISTS licenses (
			auto_id     BIGSERIAL PRIMARY KEY,
			license_key VARCHAR(64) NOT NULL UNIQUE,
			company_id  BIGINT NOT NULL DEFAULT 0,
			issued_at   TIMESTAMP NOT NULL DEFAULT now(),
			expires_at  TIMESTAMP NOT NULL,
			is_active   BOOL NOT NULL DEFAULT true
		)`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.PgPool.Exec(ctx, sql)
	if err != nil {
		log.Println("error creating licenses table    err =", err)
		return err
	}
	return nil
}

// New issues a new license for companyID valid for 6 months.
func New(companyID int64) (License, error) {
	key, err := genKey()
	if err != nil {
		return License{}, fmt.Errorf("failed to generate license key: %w", err)
	}

	l := License{
		LicenseKey: key,
		CompanyID:  companyID,
		IssuedAt:   time.Now(),
		ExpiresAt:  time.Now().AddDate(0, 6, 0),
		IsActive:   true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `INSERT INTO licenses (license_key, company_id, issued_at, expires_at, is_active)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING auto_id`

	err = database.PgPool.QueryRow(ctx, sql,
		l.LicenseKey, l.CompanyID, l.IssuedAt, l.ExpiresAt, l.IsActive,
	).Scan(&l.AutoID)
	if err != nil {
		log.Println("error inserting license    err =", err)
		return License{}, err
	}

	return l, nil
}

// Validate returns (valid, expired, error) for the active license belonging to companyID.
//   - valid=true, expired=false → license is present and in date
//   - valid=true, expired=true  → license exists but has passed its expiry
//   - valid=false               → no active license found
func Validate(companyID int64) (valid bool, expired bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `SELECT expires_at FROM licenses
			WHERE company_id = $1 AND is_active = true
			ORDER BY expires_at DESC
			LIMIT 1`

	var expiresAt time.Time
	err = database.PgPool.QueryRow(ctx, sql, companyID).Scan(&expiresAt)
	if err != nil {
		// no rows → no active license
		return false, false, nil
	}

	if time.Now().After(expiresAt) {
		return true, true, nil
	}
	return true, false, nil
}

// FetchByCompany returns the most recent active license for a company.
func FetchByCompany(companyID int64) (License, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `SELECT auto_id, license_key, company_id, issued_at, expires_at, is_active
			FROM licenses
			WHERE company_id = $1 AND is_active = true
			ORDER BY expires_at DESC
			LIMIT 1`

	l := License{}
	err := database.PgPool.QueryRow(ctx, sql, companyID).Scan(
		&l.AutoID, &l.LicenseKey, &l.CompanyID, &l.IssuedAt, &l.ExpiresAt, &l.IsActive,
	)
	if err != nil {
		return License{}, err
	}
	return l, nil
}

// Revoke deactivates all licenses for companyID.
func Revoke(companyID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.PgPool.Exec(ctx,
		`UPDATE licenses SET is_active = false WHERE company_id = $1`, companyID)
	return err
}

func genKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
