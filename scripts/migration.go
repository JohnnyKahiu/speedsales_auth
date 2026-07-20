//go:build ignore

// Run from the Login service root:
//   go run ./scripts/migration.go
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

func main() {
	// ── Load .env (host / port / dbname only) ──────────────
	_ = godotenv.Load(".env")

	host   := envOr("DB_HOST", "localhost")
	port   := envOr("DB_PORT", "5432")
	dbName := envOr("DB_NAME", "logins")

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Login Service — Database Migration")
	fmt.Printf("  DB : %s:%s/%s\n", host, port, dbName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// ── Prompt for credentials ──────────────────────────────
	dbUser := prompt("Database user: ")
	dbPass := promptPassword("Database password: ")

	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		host, port, dbName, dbUser, dbPass,
	)

	// ── Connect ─────────────────────────────────────────────
	fmt.Println()
	fmt.Print("  Connecting... ")
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		fmt.Println("FAILED")
		log.Fatalf("  connection error: %v\n", err)
	}
	defer conn.Close(ctx)
	fmt.Println("OK")
	fmt.Println()

	// ── Run migrations ──────────────────────────────────────
	steps := []struct {
		label string
		sql   string
	}{
		{"Create users table", sqlCreateUsers},
		{"Add missing columns", sqlAlterUsers},
		{"Seed default Admin", sqlSeedAdmin},
	}

	// Generate argon2 hash of "123" for the Admin seed
	adminHash, err := hashPassword("123")
	if err != nil {
		log.Fatalf("  failed to hash admin password: %v\n", err)
	}

	for _, s := range steps {
		fmt.Printf("  %-30s ", s.label+"...")
		var execErr error
		if s.label == "Seed default Admin" {
			_, execErr = conn.Exec(ctx, s.sql, adminHash)
		} else {
			_, execErr = conn.Exec(ctx, s.sql)
		}
		if execErr != nil {
			fmt.Println("FAILED")
			log.Fatalf("  %s failed: %v\n", s.label, execErr)
		}
		fmt.Println("OK")
	}

	// ── Verify ──────────────────────────────────────────────
	var exists bool
	conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = 'Admin')`).Scan(&exists)

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if exists {
		fmt.Println("  Migration complete.  Admin user is ready.")
		fmt.Println("  Default credentials:  Admin / 123")
	} else {
		fmt.Println("  Migration ran but Admin user was not found — check logs.")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// ── Helpers ───────────────────────────────────────────────

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func prompt(label string) string {
	fmt.Print("  " + label)
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	return strings.TrimSpace(sc.Text())
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, 64*1024, 1, 4,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func promptPassword(label string) string {
	fmt.Print("  " + label)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		log.Fatalf("failed to read password: %v", err)
	}
	return string(b)
}

// ── SQL ───────────────────────────────────────────────────

const sqlCreateUsers = `
CREATE TABLE IF NOT EXISTS users (
    auto_id    BIGSERIAL,
    username   VARCHAR     NOT NULL,
    first_name VARCHAR,
    last_name  VARCHAR,
    other_name VARCHAR,
    telephone  VARCHAR     NOT NULL DEFAULT '',
    email      VARCHAR     NOT NULL DEFAULT '',
    status     VARCHAR,
    company_id BIGINT      NOT NULL DEFAULT 0,
    user_class VARCHAR     NOT NULL DEFAULT 'user',
    branch     VARCHAR     NOT NULL DEFAULT '',
    department VARCHAR     NOT NULL DEFAULT '',
    role       VARCHAR     NOT NULL DEFAULT '',
    sub_role   VARCHAR,
    access_level VARCHAR,
    password   VARCHAR     NOT NULL DEFAULT '',
    stk_location VARCHAR   NOT NULL DEFAULT 'shop',
    session_id VARCHAR     NOT NULL DEFAULT '',
    till_opened TIMESTAMP,
    till        BOOL       NOT NULL DEFAULT FALSE,
    till_num    BIGINT     NOT NULL DEFAULT 0,
    device      VARCHAR(50),
    token       VARCHAR(150),
    token_date  TIMESTAMP,
    reset       BOOL       NOT NULL DEFAULT FALSE,
    profile     JSONB      NOT NULL DEFAULT '{}',

    remote_login        BOOL NOT NULL DEFAULT FALSE,
    adopt_stockcount    BOOL NOT NULL DEFAULT FALSE,
    complete_stockcount BOOL NOT NULL DEFAULT FALSE,
    pos_settings        BOOL NOT NULL DEFAULT FALSE,

    post_dispatch          BOOL NOT NULL DEFAULT FALSE,
    approve_dispatch       BOOL NOT NULL DEFAULT FALSE,
    grant_post_dispatch    BOOL NOT NULL DEFAULT FALSE,
    grant_approve_dispatch BOOL NOT NULL DEFAULT FALSE,

    post_receive          BOOL NOT NULL DEFAULT FALSE,
    approve_receive       BOOL NOT NULL DEFAULT FALSE,
    grant_post_receive    BOOL NOT NULL DEFAULT FALSE,
    grant_approve_receive BOOL NOT NULL DEFAULT FALSE,

    post_orders          BOOL NOT NULL DEFAULT FALSE,
    approve_orders       BOOL NOT NULL DEFAULT FALSE,
    grant_post_orders    BOOL NOT NULL DEFAULT FALSE,
    grant_approve_orders BOOL NOT NULL DEFAULT FALSE,

    access_sales_reports       BOOL NOT NULL DEFAULT FALSE,
    grant_access_sales_reports BOOL NOT NULL DEFAULT FALSE,

    make_sales           BOOL NOT NULL DEFAULT FALSE,
    approve_sales        BOOL NOT NULL DEFAULT FALSE,
    accept_payment       BOOL NOT NULL DEFAULT FALSE,
    grant_make_sales     BOOL NOT NULL DEFAULT FALSE,
    grant_approve_sales  BOOL NOT NULL DEFAULT FALSE,
    grant_accept_payment BOOL NOT NULL DEFAULT FALSE,

    cash_office                BOOL NOT NULL DEFAULT FALSE,
    cash_rollups               BOOL NOT NULL DEFAULT FALSE,
    approve_cash_rollups       BOOL NOT NULL DEFAULT FALSE,
    grant_cash_office          BOOL NOT NULL DEFAULT FALSE,
    grant_cash_rollups         BOOL NOT NULL DEFAULT FALSE,
    grant_approve_cash_rollups BOOL NOT NULL DEFAULT FALSE,

    sales_returns               BOOL NOT NULL DEFAULT FALSE,
    approve_sales_returns       BOOL NOT NULL DEFAULT FALSE,
    grant_sales_returns         BOOL NOT NULL DEFAULT FALSE,
    grant_approve_sales_returns BOOL NOT NULL DEFAULT FALSE,

    laybyes                    BOOL NOT NULL DEFAULT FALSE,
    approve_credit_sales       BOOL NOT NULL DEFAULT FALSE,
    grant_laybyes              BOOL NOT NULL DEFAULT FALSE,
    grant_approve_credit_sales BOOL NOT NULL DEFAULT FALSE,

    activate_mpesa       BOOL NOT NULL DEFAULT FALSE,
    grant_activate_mpesa BOOL NOT NULL DEFAULT FALSE,

    post_cheques          BOOL NOT NULL DEFAULT FALSE,
    approve_cheques       BOOL NOT NULL DEFAULT FALSE,
    grant_post_cheques    BOOL NOT NULL DEFAULT FALSE,
    grant_approve_cheques BOOL NOT NULL DEFAULT FALSE,

    create_users       BOOL NOT NULL DEFAULT FALSE,
    delete_users       BOOL NOT NULL DEFAULT FALSE,
    grant_create_users BOOL NOT NULL DEFAULT FALSE,
    grant_delete_users BOOL NOT NULL DEFAULT FALSE,

    price_change         BOOL NOT NULL DEFAULT FALSE,
    ammend_invoice       BOOL NOT NULL DEFAULT FALSE,
    grant_price_change   BOOL NOT NULL DEFAULT FALSE,
    grant_ammend_invoice BOOL NOT NULL DEFAULT FALSE,

    create_stock        BOOL NOT NULL DEFAULT FALSE,
    link_stock          BOOL NOT NULL DEFAULT FALSE,
    audit_stock         BOOL NOT NULL DEFAULT FALSE,
    complete_stock_take BOOL NOT NULL DEFAULT FALSE,
    grant_create_stock  BOOL NOT NULL DEFAULT FALSE,
    grant_audit_stock   BOOL NOT NULL DEFAULT FALSE,

    accounts              BOOL NOT NULL DEFAULT FALSE,
    aprove_accounts       BOOL NOT NULL DEFAULT FALSE,
    grant_accounts        BOOL NOT NULL DEFAULT FALSE,
    grant_aprove_accounts BOOL NOT NULL DEFAULT FALSE,

    ledger             BOOL NOT NULL DEFAULT FALSE,
    recon_ledger       BOOL NOT NULL DEFAULT FALSE,
    grant_ledger       BOOL NOT NULL DEFAULT FALSE,
    grant_recon_ledger BOOL NOT NULL DEFAULT FALSE,

    create_scard       BOOL NOT NULL DEFAULT FALSE,
    grant_create_scard BOOL NOT NULL DEFAULT FALSE,

    produce       BOOL NOT NULL DEFAULT FALSE,
    grant_produce BOOL NOT NULL DEFAULT FALSE,

    CONSTRAINT users_username_key PRIMARY KEY (username)
)`

const sqlAlterUsers = `
ALTER TABLE users ADD COLUMN IF NOT EXISTS accept_payment         BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_accept_payment   BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS delete_users           BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_delete_users     BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS adopt_stockcount       BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS complete_stockcount    BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS remote_login           BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS cash_office            BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_cash_office      BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS activate_mpesa         BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_activate_mpesa   BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS post_cheques           BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS approve_cheques        BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_post_cheques     BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_approve_cheques  BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS ammend_invoice         BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_ammend_invoice   BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS produce                BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_produce          BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile                JSONB NOT NULL DEFAULT '{}'`

const sqlSeedAdmin = `
INSERT INTO users (
    username, first_name, user_class, password, reset,
    remote_login, adopt_stockcount, complete_stockcount, pos_settings,
    post_dispatch, approve_dispatch, grant_post_dispatch, grant_approve_dispatch,
    post_receive, approve_receive, grant_post_receive, grant_approve_receive,
    post_orders, approve_orders, grant_post_orders, grant_approve_orders,
    access_sales_reports, grant_access_sales_reports,
    make_sales, approve_sales, accept_payment,
    grant_make_sales, grant_approve_sales, grant_accept_payment,
    cash_office, cash_rollups, approve_cash_rollups,
    grant_cash_office, grant_cash_rollups, grant_approve_cash_rollups,
    sales_returns, approve_sales_returns,
    grant_sales_returns, grant_approve_sales_returns,
    laybyes, approve_credit_sales, grant_laybyes, grant_approve_credit_sales,
    activate_mpesa, grant_activate_mpesa,
    post_cheques, approve_cheques, grant_post_cheques, grant_approve_cheques,
    create_users, delete_users, grant_create_users, grant_delete_users,
    price_change, ammend_invoice, grant_price_change, grant_ammend_invoice,
    create_stock, link_stock, audit_stock, complete_stock_take,
    grant_create_stock, grant_audit_stock,
    accounts, aprove_accounts, grant_accounts, grant_aprove_accounts,
    ledger, recon_ledger, grant_ledger, grant_recon_ledger,
    create_scard, grant_create_scard,
    produce, grant_produce
)
SELECT
    'Admin', 'Administrator', 'super user', $1, FALSE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE,
    TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE,
    TRUE, TRUE,
    TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE,
    TRUE, TRUE
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'Admin')`
