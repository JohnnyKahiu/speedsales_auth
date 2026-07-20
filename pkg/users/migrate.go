package users

import (
	"context"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_login/pkg/database"
)

// Migrate ensures the users table schema is up to date and the Admin user
// exists with full grant rights. Safe to run on both fresh and existing databases.
func Migrate() error {
	if err := genUserTable(); err != nil {
		log.Println("migrate: schema error    err =", err)
		return err
	}
	if err := upsertAdminUser(); err != nil {
		log.Println("migrate: admin upsert error    err =", err)
		return err
	}
	return nil
}

// upsertAdminUser creates the Admin user if it does not exist, then always
// sets every grant_* right to true so new permissions are picked up after migration.
func upsertAdminUser() error {
	config := &PasswordConfig{Time: 1, Memory: 64 * 1024, Threads: 4, KeyLen: 32}
	hashedPwd, err := GeneratePassword(config, "123")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Create Admin only if username doesn't already exist.
	insertSQL := `
		INSERT INTO users (username, first_name, user_class, password, reset)
		VALUES ('Admin', 'Administrator', 'super user', $1, true)
		ON CONFLICT (username) DO NOTHING`

	if _, err = database.PgPool.Exec(ctx, insertSQL, hashedPwd); err != nil {
		log.Println("migrate: failed to insert admin user    err =", err)
		return err
	}

	// Always set all grant rights and operational flags to true on the Admin row.
	updateSQL := `
		UPDATE users SET
			remote_login               = true,
			pos_settings               = true,
			adopt_stockcount           = true,
			complete_stockcount        = true,
			complete_stock_take        = true,
			post_dispatch              = true,
			approve_dispatch           = true,
			grant_post_dispatch        = true,
			grant_approve_dispatch     = true,
			post_receive               = true,
			approve_receive            = true,
			grant_post_receive         = true,
			grant_approve_receive      = true,
			post_orders                = true,
			approve_orders             = true,
			grant_post_orders          = true,
			grant_approve_orders       = true,
			access_sales_reports       = true,
			grant_access_sales_reports = true,
			make_sales                 = true,
			approve_sales              = true,
			accept_payment             = true,
			grant_make_sales           = true,
			grant_approve_sales        = true,
			grant_accept_payment       = true,
			cash_office                = true,
			cash_rollups               = true,
			approve_cash_rollups       = true,
			grant_cash_office          = true,
			grant_cash_rollups         = true,
			grant_approve_cash_rollups = true,
			sales_returns              = true,
			approve_sales_returns      = true,
			grant_sales_returns        = true,
			grant_approve_sales_returns= true,
			laybyes                    = true,
			approve_credit_sales       = true,
			grant_laybyes              = true,
			grant_approve_credit_sales = true,
			activate_mpesa             = true,
			grant_activate_mpesa       = true,
			post_cheques               = true,
			approve_cheques            = true,
			grant_post_cheques         = true,
			grant_approve_cheques      = true,
			create_users               = true,
			delete_users               = true,
			grant_create_users         = true,
			grant_delete_users         = true,
			price_change               = true,
			ammend_invoice             = true,
			grant_price_change         = true,
			grant_ammend_invoice       = true,
			create_stock               = true,
			link_stock                 = true,
			audit_stock                = true,
			grant_create_stock         = true,
			grant_audit_stock          = true,
			grant_complete_stockcount  = true,
			accounts                   = true,
			aprove_accounts            = true,
			grant_accounts             = true,
			grant_aprove_accounts      = true,
			ledger                     = true,
			recon_ledger               = true,
			grant_ledger               = true,
			grant_recon_ledger         = true,
			create_scard               = true,
			grant_create_scard         = true,
			produce                    = true,
			grant_produce              = true
		WHERE username = 'Admin'`

	if _, err = database.PgPool.Exec(ctx, updateSQL); err != nil {
		log.Println("migrate: failed to update admin grants    err =", err)
		return err
	}
	return nil
}
