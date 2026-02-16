package users

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	random "math/rand"
	"os"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/JohnnyKahiu/speedsales_login/pkg/database"
	"github.com/JohnnyKahiu/speedsales_login/pkg/variables"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

// PasswordConfig password configuration
type PasswordConfig struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

// Token holds token parameters
type Token struct {
	Authorized bool      `json:"authorized"`
	Username   string    `json:"username"`
	Rights     Users     `json:"rights"`
	Expiry     time.Time `json:"exp"`
}

var mySigningKey = os.Getenv("JWT_KEY")

// BytesToString converts bytes array into a string
func BytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{bh.Data, bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

// CreateUser creaates a new user
func (args *Users) CreateUser(password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if args.Username == "" {
		return fmt.Errorf("failed to create user username is null")
	}

	// make user as default user class
	if args.UserClass == "" {
		args.UserClass = "user"
	}

	args.Reset = true

	sql := `INSERT INTO users(first_name, last_name, telephone, email, username, branch, company_id, user_class, pos_settings, 
				post_dispatch, approve_dispatch, grant_post_dispatch, grant_approve_dispatch, 
				post_receive, approve_receive, grant_post_receive, grant_approve_receive,
				post_orders, approve_orders , grant_post_orders, grant_approve_orders,
				make_sales, approve_sales, grant_make_Sales, grant_approve_sales,
				sales_returns,  approve_sales_returns, grant_sales_returns, grant_approve_sales_returns,
				grant_cash_rollups, cash_rollups, approve_cash_rollups, grant_approve_cash_rollups,
				laybyes, approve_credit_sales, grant_laybyes, grant_approve_credit_sales,
				ledger, recon_ledger, create_scard, 
				accounts, grant_accounts, aprove_accounts, grant_aprove_accounts, 
				create_users, delete_users, create_stock,
				access_sales_reports, stk_location, reset, password)
			VALUES($1, $2, $3, $4, $5, $6, $7, 
					$8, $9, $10, $11, 
					$12, $13, $14, $15, 
					$16, $17, $18, $19, 
					$20, $21, $22, $23, 
					$24, $25, $26, $27, 
					$28, $29, $30, $31, 
					$32, $33, $34, $35, 
					$36, $37, $38, 
					$39, $40, $41, 
					$42, $43, $44, 
					$45, $46, $47, 
					$48, $49, $50, $51 )`

	// fmt.Println(sql)
	_, err := database.PgPool.Exec(ctx, sql, args.FirstName, args.LastName, args.Telephone, args.Email, args.Username, args.Branch, args.CompanyID, args.UserClass, args.PosSettings,
		args.PostDispatch, args.ApproveDispatch, args.GrantPostDispatch, args.GrantApproveDispatch,
		args.PostReceive, args.ApproveReceive, args.GrantPostReceive, args.GrantApproveReceive,
		args.PostOrders, args.ApproveOrders, args.GrantPostOrders, args.GrantApproveOrders,
		args.MakeSales, args.ApproveSales, args.GrantMakeSales, args.GrantApproveSales,
		args.SalesReturns, args.ApproveSalesReturns, args.GrantSalesReturns, args.GrantApproveSalesReturns,
		args.CashRollups, args.GrantCashRollups, args.ApproveCashRollups, args.GrantApproveCashRollups,
		args.Laybyes, args.ApproveCreditSales, args.GrantLaybyes, args.GrantApproveCreditSales,
		args.Ledger, args.ReconLedger, args.CreateScard,
		args.Accounts, args.GrantAccounts, args.AproveAccounts, args.GrantAproveAccounts,
		args.CreateUsers, args.DeleteUsers, args.CreateStock,
		args.AccessSalesReports, args.StkLocation, args.Reset, password)

	if err != nil {
		log.Println("\n Error creating user ", err)
		return err
	}

	return nil
}

// FetchUser gets user details from database
func FetchUser(ctx context.Context, username string) (Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Printf("\nfetching user for username: '%v'", username)
	// create a values slice and scan values into it
	var values Users

	// create a sql fetch_statement
	sql := `SELECT first_name, last_name, 
				username, coalesce(password, '123'), branch, coalesce(company_id, 0), coalesce(user_class, 'user'), pos_settings, 
				coalesce(post_dispatch, false), coalesce(approve_dispatch, false), coalesce(grant_post_dispatch, false), coalesce(grant_approve_dispatch, false), 
				coalesce(post_receive, false), coalesce(approve_receive, false), coalesce(grant_post_receive, false), coalesce(grant_approve_receive, false),
				coalesce(post_orders, false), coalesce(approve_orders, false) , coalesce(grant_post_orders, false), coalesce(grant_approve_orders, false) ,
				produce, grant_produce, 
				coalesce(make_sales, false), coalesce(approve_sales, false), coalesce(accept_payment, false), coalesce(grant_make_Sales, false), coalesce(grant_approve_sales, false), coalesce(grant_accept_payment, false),
				coalesce(sales_returns, false),  coalesce(approve_sales_returns, false), coalesce(grant_sales_returns, false), coalesce(grant_approve_sales_returns, false),
				coalesce(grant_cash_rollups, false), coalesce(cash_rollups, false), coalesce(approve_cash_rollups, false), coalesce(grant_approve_cash_rollups, false),
				coalesce(laybyes, false), coalesce(approve_credit_sales, false), coalesce(grant_laybyes, false), coalesce(grant_approve_credit_sales, false),
				coalesce(ledger, false), coalesce(recon_ledger, false), coalesce(access_sales_reports, false),
				coalesce(activate_mpesa, false), coalesce(grant_activate_mpesa, false), coalesce(cash_office, false), coalesce(grant_cash_office, false),
				coalesce(accounts, false), coalesce(aprove_accounts, false), coalesce(complete_stock_take, false),
				coalesce(create_scard, false), coalesce(grant_create_scard, false), coalesce(create_stock, false), coalesce(link_stock, false), coalesce(price_change, false), coalesce(grant_price_change, false), coalesce(stk_location, 'shop'),
				till_num, reset,
				coalesce(cast(token as varchar(20)), 'NULL') token, 	coalesce(token_date, '2006-01-01'),
				coalesce(adopt_stockcount, false), coalesce(create_users, false), coalesce(ammend_invoice, false), coalesce(grant_ammend_invoice, false)
			FROM users WHERE username = $1`

	// run query
	rows, err := database.PgPool.Query(ctx, sql, username)
	if err != nil {
		fmt.Println("error fetching user ", err.Error())
		return values, err
	}
	defer rows.Close()

	for rows.Next() {
		// scan rows into user_details
		rows.Scan(&values.FirstName, &values.LastName,
			&values.Username, &values.password, &values.Branch, &values.CompanyID, &values.UserClass, &values.PosSettings,
			&values.PostDispatch, &values.ApproveDispatch, &values.GrantPostDispatch, &values.GrantApproveDispatch,
			&values.PostReceive, &values.ApproveReceive, &values.GrantPostReceive, &values.GrantApproveReceive,
			&values.PostOrders, &values.ApproveOrders, &values.GrantPostOrders, &values.GrantApproveOrders,
			&values.Produce, &values.GrantProduce,
			&values.MakeSales, &values.ApproveSales, &values.AcceptPayment, &values.GrantMakeSales, &values.GrantApproveSales, &values.GrantAcceptPayment,
			&values.SalesReturns, &values.ApproveSalesReturns, &values.GrantSalesReturns, &values.GrantApproveSalesReturns,
			&values.GrantCashRollups, &values.CashRollups, &values.ApproveCashRollups, &values.GrantApproveCashRollups,
			&values.Laybyes, &values.ApproveCreditSales, &values.GrantLaybyes, &values.GrantApproveCreditSales,
			&values.Ledger, &values.ReconLedger, &values.AccessSalesReports,
			&values.ActivateMpesa, &values.GrantActivateMpesa, &values.CashOffice, &values.GrantCashOffice,
			&values.Accounts, &values.AproveAccounts, &values.CompleteStockTake,
			&values.CreateScard, &values.GrantCreateScard, &values.CreateStock, &values.LinkStock, &values.PriceChange, &values.GrantPriceChange, &values.StkLocation,
			&values.TillNum, &values.Reset,
			&values.Token, &values.TokenDate,
			&values.AdoptStockcount, &values.CreateUsers, &values.AmmendInvoice, &values.GrantAmmendInvoice)
	}

	// return values
	return values, nil
}

func UpdateTill(ctx context.Context, username, till_no string) error {
	sql := `UPDATE users SET till_num = $1 WHERE username = $2`

	_, err := database.PgPool.Exec(ctx, sql, till_no, username)
	if err != nil {
		log.Println("failed to update till_num to users    err =", err)
		return err
	}
	return nil
}

// UpdateUser updates current user
func UpdateUser(rights Users, args map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sql := "status = 'ACTIVE'"

	if rights.CreateUsers {
		if rights.UserClass == "super user" {

			for key, val := range args {
				key = strings.Replace(fmt.Sprintf("%v", key), "'", "|| chr(39) ||", -1)
				val = strings.Replace(fmt.Sprintf("%v", val), "'", "|| chr(39) ||", -1)

				if key != "first_name" && key != "last_name" && key != "last_name" && key != "other_name" && key != "telephone" && key != "status" && key != "username" && key != "email" && key != "branch" && key != "department" && key != "stk_location" {
					sql += fmt.Sprintf(", %v = %v", key, val)
				} else {
					sql += fmt.Sprintf(", %v = '%v'", key, val)
				}
			}
		} else {
			// update grant rights only if super user

			if val, ok := args["first_name"]; ok {
				sql += fmt.Sprintf(", first_name = '%v'", strings.Replace(fmt.Sprintf("%v", val), "'", "|| chr(39) ||", -1))
			}
			if val, ok := args["last_name"]; ok {
				sql += fmt.Sprintf(", last_name = '%v'", strings.Replace(fmt.Sprintf("%v", val), "'", "|| chr(39) ||", -1))
			}
			if val, ok := args["telephone"]; ok {
				sql += fmt.Sprintf(", telephone = '%v'", strings.Replace(fmt.Sprintf("%v", val), "'", "|| chr(39) ||", -1))
			}
			if val, ok := args["email"]; ok {
				sql += fmt.Sprintf(", email = '%v'", strings.Replace(fmt.Sprintf("%v", val), "'", "|| chr(39) ||", -1))
			}
			if val, ok := args["company_id"]; ok {
				sql += fmt.Sprintf(", company_id = '%v'", strings.Replace(fmt.Sprintf("%v", val), "'", "|| chr(39) ||", -1))
			}
			if val, ok := args["branch"]; ok {
				// fmt.Println("branch exists")
				sql += fmt.Sprintf(", branch = '%v'", strings.Replace(fmt.Sprintf("%v", val), "'", "|| chr(39) ||", -1))
			}
			if val, ok := args["department"]; ok {
				sql += fmt.Sprintf(", department = '%v'", strings.Replace(fmt.Sprintf("%v", val), "'", "|| chr(39) ||", -1))
			}
			if val, ok := args["stk_location"]; ok {
				sql += fmt.Sprintf(", stk_location = '%v'", strings.Replace(fmt.Sprintf("%v", val), "'", "|| chr(39) ||", -1))
			}

			// ==================== Sales rights ==========================================
			// update make sales if can grant
			if rights.GrantMakeSales || rights.UserClass == "super user" {
				if val, ok := args["make_sales"]; ok {
					sql += fmt.Sprintf(", make_sales = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_make_sales"]; ok {
					sql += fmt.Sprintf(", grant_make_sales = %v", val)
				}
			}

			// update accept payment if can grant
			if rights.GrantAcceptPayment || rights.UserClass == "super user" {
				if val, ok := args["accept_payment"]; ok {
					sql += fmt.Sprintf(", accept_payment = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_accept_payment"]; ok {
					sql += fmt.Sprintf(", grant_accept_payment = %v", val)
				}
			}

			// update approve sales if can grant
			if rights.GrantApproveSales || rights.UserClass == "super user" {
				if val, ok := args["approve_sales"]; ok {
					sql += fmt.Sprintf(", make_sales = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_approve_sales"]; ok {
					sql += fmt.Sprintf(", grant_approve_sales = %v", val)
				}
			}

			// update create_scard
			if rights.GrantCreateScard || rights.UserClass == "super user" {
				if val, ok := args["create_scard"]; ok {
					sql += fmt.Sprintf(", create_scard = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_create_scard"]; ok {
					sql += fmt.Sprintf(", grant_create_scard = %v", val)
				}
			}

			// update laybye sales if can grant
			if rights.GrantLaybyes || rights.UserClass == "super user" {
				if val, ok := args["laybyes"]; ok {
					sql += fmt.Sprintf(", laybyes = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_laybyes"]; ok {
					sql += fmt.Sprintf(", grant_laybyes = %v", val)
				}
			}

			// update laybye sales if can grant
			if rights.GrantApproveCreditSales || rights.UserClass == "super user" {
				if val, ok := args["approve_credit_sales"]; ok {
					sql += fmt.Sprintf(", approve_credit_sales = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_approve_credit_sales"]; ok {
					sql += fmt.Sprintf(", grant_approve_credit_sales = %v", val)
				}
			}

			// update sales returns if can grant
			if rights.GrantSalesReturns || rights.UserClass == "super user" {
				if val, ok := args["sales_returns"]; ok {
					sql += fmt.Sprintf(", sales_returns = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_sales_returns"]; ok {
					sql += fmt.Sprintf(", grant_sales_returns = %v", val)
				}
			}

			// update approve sales returns if can grant
			if rights.GrantApproveSalesReturns || rights.UserClass == "super user" {
				if val, ok := args["approve_sales_returns"]; ok {
					sql += fmt.Sprintf(", approve_sales_returns = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_approve_sales_returns"]; ok {
					sql += fmt.Sprintf(", grant_approve_sales_returns = %v", val)
				}
			}

			// update activate mpesa if can grant
			if rights.GrantActivateMpesa || rights.UserClass == "super user" {
				if val, ok := args["activate_mpesa"]; ok {
					sql += fmt.Sprintf(", activate_mpesa = %v", val)
				}
			}

			// update cash rollups if can grant
			if rights.GrantCashRollups || rights.UserClass == "super user" {
				if val, ok := args["cash_rollups"]; ok {
					sql += fmt.Sprintf(", cash_rollups = %v", val)
				}
			}

			if rights.GrantCashOffice || rights.UserClass == "super user" {
				if val, ok := args["cash_office"]; ok {
					sql += fmt.Sprintf(", cash_office = %v", val)
				}
			}

			if rights.UserClass == "super user" {
				if val, ok := args["grant_cash_rollups"]; ok {
					sql += fmt.Sprintf(", grant_cash_rollups = %v", val)
				}
			}

			// update access sales reports if can grant
			if rights.GrantAccessSalesReports || rights.UserClass == "super user" {
				if val, ok := args["access_sales_reports"]; ok {
					sql += fmt.Sprintf(", access_sales_reports = %v", val)
				}
			}

			// ==================== Aquisition rights ==========================================
			// update approve cash rollups if can grant
			if rights.GrantApproveCashRollups || rights.UserClass == "super user" {
				if val, ok := args["approve_cash_rollups"]; ok {
					sql += fmt.Sprintf(", approve_cash_rollups = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_approve_cash_rollups"]; ok {
					sql += fmt.Sprintf(", grant_approve_cash_rollups = %v", val)
				}
			}

			// update post dispatch if can grant
			if rights.GrantPostDispatch || rights.UserClass == "super user" {
				if val, ok := args["post_dispatch"]; ok {
					sql += fmt.Sprintf(", post_dispatch = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_approve_cash_rollups"]; ok {
					sql += fmt.Sprintf(", grant_approve_cash_rollups = %v", val)
				}
			}

			// update approve dispatch if can grant
			if rights.GrantApproveDispatch || rights.UserClass == "super user" {
				if val, ok := args["approve_dispatch"]; ok {
					sql += fmt.Sprintf(", approve_dispatch = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_approve_cash_rollups"]; ok {
					sql += fmt.Sprintf(", grant_approve_cash_rollups = %v", val)
				}
			}

			// update post receive rights if can grant
			if rights.GrantPostReceive || rights.UserClass == "super user" {
				if val, ok := args["post_receive"]; ok {
					sql += fmt.Sprintf(", post_receive = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_post_receive"]; ok {
					sql += fmt.Sprintf(", grant_post_receive = %v", val)
				}
			}

			// update approve receive if can grant
			if rights.GrantApproveReceive || rights.UserClass == "super user" {
				if val, ok := args["approve_receive"]; ok {
					sql += fmt.Sprintf(", approve_receive = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_approve_receive"]; ok {
					sql += fmt.Sprintf(", grant_approve_receive = %v", val)
				}
			}

			// update post orders if can grant
			if rights.GrantPostOrders || rights.UserClass == "super user" {
				if val, ok := args["post_orders"]; ok {
					sql += fmt.Sprintf(", post_orders = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_post_orders"]; ok {
					sql += fmt.Sprintf(", grant_post_orders = %v", val)
				}
			}

			// update approve orders if can grant
			if rights.GrantApproveOrders || rights.UserClass == "super user" {
				if val, ok := args["approve_orders"]; ok {
					sql += fmt.Sprintf(", approve_orders = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_approve_orders"]; ok {
					sql += fmt.Sprintf(", grant_approve_orders = %v", val)
				}
			}

			if rights.GrantProduce || rights.UserClass == "super user" {
				if val, ok := args["produce"]; ok {
					sql += fmt.Sprintf(", produce = %v", val)
				}
			}
			if rights.UserClass == "super user" {
				if val, ok := args["grant_produce"]; ok {
					sql += fmt.Sprintf(", grant_produce = %v", val)
				}
			}

			// ==================== Stock rights ==========================================
			// update price change if can grant
			if rights.GrantPriceChange || rights.UserClass == "super user" {
				if val, ok := args["price_change"]; ok {
					sql += fmt.Sprintf(", price_change = %v", val)
				}
			}

			// update create stock, link stock if can grant create stock
			if rights.GrantCreateStock || rights.UserClass == "super user" {
				if val, ok := args["create_stock"]; ok {
					sql += fmt.Sprintf(", create_stock = %v", val)
				}
				if val, ok := args["link_stock"]; ok {
					sql += fmt.Sprintf(", link_stock = %v", val)
				}
			}

			// ==================== User Management and Accounts rights ==========================================
			// update create user if can grant
			if rights.GrantCreateUsers || rights.UserClass == "super user" {
				if val, ok := args["create_users"]; ok {
					sql += fmt.Sprintf(", create_users = %v", val)
				}
			}

			// update Ledger, access accounts, post_cheques  if can grant Accounts
			if rights.GrantAccounts || rights.UserClass == "super user" {
				if val, ok := args["accounts"]; ok {
					sql += fmt.Sprintf(", accounts = %v", val)
				}
				if val, ok := args["ledger"]; ok {
					sql += fmt.Sprintf(", ledger = %v", val)
				}
				if val, ok := args["post_cheques"]; ok {
					sql += fmt.Sprintf(", post_cheques = %v", val)
				}
				if val, ok := args["adopt_stockcount"]; ok {
					sql += fmt.Sprintf(", adopt_stockcount = %v", val)
				}
			}

			// update recon_ledger, aprove_accounts, approve_cheques  if can grant approve accounts
			if rights.GrantAproveAccounts || rights.UserClass == "super user" {
				if val, ok := args["recon_ledger"]; ok {
					sql += fmt.Sprintf(", recon_ledger = %v", val)
				}
				if val, ok := args["aprove_accounts"]; ok {
					sql += fmt.Sprintf(", aprove_accounts = %v", val)
				}
				if val, ok := args["approve_cheques"]; ok {
					sql += fmt.Sprintf(", approve_cheques = %v", val)
				}
			}
		}
	}
	fmt.Println("username =", args["username"])

	if val, ok := args["username"]; ok {
		username := strings.Replace(fmt.Sprintf("%v", val), "'", "||chr(39)||", -1)

		SQL := fmt.Sprintf("UPDATE users SET %v WHERE username = '%v'", sql, username)
		fmt.Printf("\tsql = %v\n", SQL)

		_, err := database.PgPool.Exec(ctx, SQL)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("username is not provided")
	}

	return nil
}

// FetchAllActiveUsers fetches all active users
func FetchAllActiveUsers(branch string) ([]Users, error) {
	brCon := " "
	if branch != "" && branch != "all" {
		brCon = fmt.Sprintf("WHERE branch = '%v'", strings.Replace(branch, "'", "|| chr(39) ||", -1))
	}

	// create a sql fetch_statement
	sql := fmt.Sprintf(`
			SELECT first_name, last_name, 
				username, branch, coalesce(company_id, 0), coalesce(user_class, 'user'), pos_settings, 
				post_dispatch, approve_dispatch, grant_post_dispatch, grant_approve_dispatch, 
				post_receive, approve_receive, grant_post_receive, grant_approve_receive,
				post_orders, approve_orders , grant_post_orders, grant_approve_orders,
				produce, grant_produce,
				make_sales, approve_sales, grant_make_Sales, grant_approve_sales,
				sales_returns,  approve_sales_returns, grant_sales_returns, grant_approve_sales_returns,
				grant_cash_rollups, cash_rollups, approve_cash_rollups, grant_approve_cash_rollups,
				laybyes, approve_credit_sales, grant_laybyes, grant_approve_credit_sales,
				ledger, recon_ledger, access_sales_reports,
				accounts, aprove_accounts, complete_stock_take,
				create_scard, create_stock, stk_location, reset,
       			adopt_stockcount, create_users
			FROM users %v
			ORDER BY username ASC`, brCon)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// run query
	rows, err := database.PgPool.Query(ctx, sql)
	if err != nil {
		log.Println("error, users db query error     err =", err)
		return nil, err
	}
	defer rows.Close()

	var vals []Users
	for rows.Next() {
		var values Users
		// scan rows into user_details
		err := rows.Scan(&values.FirstName, &values.LastName,
			&values.Username, &values.Branch, &values.CompanyID, &values.UserClass, &values.PosSettings,
			&values.PostDispatch, &values.ApproveDispatch, &values.GrantPostDispatch, &values.GrantApproveDispatch,
			&values.PostReceive, &values.ApproveReceive, &values.GrantPostReceive, &values.GrantApproveReceive,
			&values.PostOrders, &values.ApproveOrders, &values.GrantPostOrders, &values.GrantApproveOrders,
			&values.Produce, &values.GrantProduce,
			&values.MakeSales, &values.ApproveSales, &values.GrantMakeSales, &values.GrantApproveSales,
			&values.SalesReturns, &values.ApproveSalesReturns, &values.GrantSalesReturns, &values.GrantApproveSalesReturns,
			&values.GrantCashRollups, &values.CashRollups, &values.ApproveCashRollups, &values.GrantApproveCashRollups,
			&values.Laybyes, &values.ApproveCreditSales, &values.GrantLaybyes, &values.GrantApproveCreditSales,
			&values.Ledger, &values.ReconLedger, &values.AccessSalesReports,
			&values.Accounts, &values.AproveAccounts, &values.CompleteStockTake,
			&values.CreateScard, &values.CreateStock, &values.StkLocation, &values.Reset,
			&values.AdoptStockcount, &values.CreateUsers)

		if err != nil {
			fmt.Println(err)
		}

		vals = append(vals, values)
	}

	// return values
	return vals, nil
}

// CreateToken generates user's approve token
func CreateToken(username string, duration int) error {
	// make default duration to 30 days
	duration = 30

	// generate random 8 digit number
	ind0 := random.Intn(9)
	ind1 := random.Intn(9)
	ind2 := random.Intn(9)
	ind3 := random.Intn(9)
	ind4 := random.Intn(9)
	ind5 := random.Intn(9)
	ind6 := random.Intn(9)
	ind7 := random.Intn(9)
	ind8 := random.Intn(9)

	token := fmt.Sprintf("%v%v%v%v%v%v%v%v%v", ind0, ind1, ind2, ind3, ind4, ind5, ind6, ind7, ind8)
	// fmt.Println("signature token = ", token)

	tokenDate := time.Now()
	tokenDate = tokenDate.AddDate(0, 0, duration)

	// update token and expiry date
	sql := `UPDATE users
			SET
				token = $1,
				token_date = $2
			WHERE 
				username = $3
			`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// run sql
	_, err := database.PgPool.Exec(ctx, sql, token, tokenDate, username)
	if err != nil {
		return err
	}

	return nil
}

// ResetPassword sets new password for specified user
func ResetPassword(password, username string) error {
	if username == "" {
		return fmt.Errorf("username is not provided")
	}
	if password == "" {
		return fmt.Errorf("password cannot be null")
	}

	fmt.Println("\n\t username =", username)
	// sql string
	sql := `UPDATE users 
			SET 
				password = $1, 
				reset = false 
			WHERE username = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// run sql statement
	_, err := database.PgPool.Exec(ctx, sql, password, username)
	if err != nil {
		log.Printf("\n error while updating user\n\t %v\n", err)
		return err
	}

	fmt.Println("\t reset password successful")
	return nil
}

// GenToken returns a token (not so sure token or key)
func GenToken(chars int) string {
	token := ""
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for i := 0; i < chars; i++ {
		token += fmt.Sprintf("%v", random.Intn(9))
	}
	return token
}

// SetSecurityToken updates new api ket
func SetSecurityToken(email, username string) error {
	// sql string
	sql := "UPDATE users SET security_token = $1 WHERE username = $2"

	// generate token
	token := GenToken(30)
	// fmt.Printf("\ngen token = %v\n", token)

	// create a web stocken
	jwt := make(map[string]string)
	jwt["username"] = username
	jwt["key"] = token
	jwt["expiry"] = fmt.Sprintf("%v", time.Now().Add(time.Hour*time.Duration(12)))

	// create a json string
	jwtStr, _ := json.Marshal(jwt)

	// fmt.Printf("\n web token = %v\n key =%v\n", string(jwtStr), token)

	// define argon password configuration
	config := &PasswordConfig{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	}
	// hash token
	hash, err := GeneratePassword(config, string(jwtStr))
	if err != nil {

	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//  run statement
	rows, err := database.PgPool.Query(ctx, sql, hash, username)
	if err != nil {
		return err
	}
	defer rows.Close()

	return nil
}

// GeneratePassword is used to generate a new password hash for storing
func GeneratePassword(c *PasswordConfig, password string) (string, error) {

	// Generate a Salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, c.Time, c.Memory, c.Threads, c.KeyLen)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	format := "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	full := fmt.Sprintf(format, argon2.Version, c.Memory, c.Time, c.Threads, b64Salt, b64Hash)
	return full, nil
}

// ComparePassword is used to compare a user-inputted password to a hash  if they match.
func ComparePassword(username, password string) (bool, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	/*
		compares hashed passwords and returns: (bool, bool, error)
			bool  := true or false if user is authenticated
			bool  := true or false if user is reset or not
			error := error thrown from running statement
	*/
	// search user
	userDetails, err := FetchUser(ctx, username)
	fmt.Println("\nhashed_pass =", userDetails.password)
	fmt.Println("password =", password)

	if err != nil {
		return false, false, err
	}
	if userDetails.Reset {
		return true, true, nil
	} else if userDetails.password == "" && userDetails.Reset {
		return true, true, nil
	} else if userDetails.password == "" && userDetails.UserClass == "replication" {
		return true, false, nil
	}
	if userDetails.password == "" {
		return false, false, nil
	}
	hash := userDetails.password

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

// DecodeHash unhashes a hash text to plain string
func DecodeHash(hash string) (string, error) {
	parts := strings.Split(hash, "$")

	c := &PasswordConfig{}
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &c.Memory, &c.Time, &c.Threads)
	if err != nil {
		return "", err
	}

	_, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return "", err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return "", err
	}

	// fmt.Printf("\nDecoded Hash = %v\n", BytesToString(decodedHash))
	return BytesToString(decodedHash), nil
}

// GenerateJWT generates security token [json web token]
func GenerateJWT(username string) (string, Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	// fetch user details
	userDetails, err := FetchUser(ctx, username)

	t := time.Now().Local().Add(time.Hour * time.Duration(22))

	sessionID := ""
	if userDetails.UserClass != "replication" {
		sessionID = CreateSessKey(username)
	} else {
		sessionID = CreateSessKey("")
	}

	var sessionIDs []string
	if userDetails.MakeSales {
		sessionIDs = append(sessionIDs, sessionID)
	} else {
		sIds, _ := getSessionID(username)
		if len(sIds) >= 4 {
			for i := 1; i < len(sIds); i++ {
				sessionIDs = append(sessionIDs, sIds[i])
			}
			sessionIDs = append(sessionIDs, sessionID)
		} else {
			sessionIDs = sIds
			sessionIDs = append(sessionIDs, sessionID)
		}
	}
	sessionIDstr, _ := json.Marshal(sessionIDs)

	// fmt.Println("\n\n session id =", string(sessionIDstr), "\n\n .")

	claims["authorized"] = true
	claims["session"] = sessionID
	claims["rights"] = userDetails
	claims["exp"] = fmt.Sprintf("%d-%02d-%02d %02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	claims["username"] = username

	var user Users
	if variables.Cache {
		// log session id to redis
		database.RdbCon.Set(ctx, username, string(sessionIDstr), 0)
	} else {
		fmt.Printf("\tusername = %v \t session_id = %v\n", username, string(sessionIDstr))
		err = setSessionID(username, string(sessionIDstr))
		if err != nil {
			log.Println("failed to save sessionID error =", err)
			return "", user, err
		}
	}

	// fmt.Printf("\n\nCreate expiry date = %v\n", claims["exp"])
	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		return "", user, err
	}

	return tokenString, userDetails, nil
}

// GenLoginToken
func GenLoginToken(userDetails Users) (string, Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	t := time.Now().Local().Add(time.Hour * time.Duration(22))

	// fmt.Println("\n\n session id =", string(sessionIDstr), "\n\n .")
	username := userDetails.Username

	sessionIDs := CreateSessKey(username)
	claims["authorized"] = true
	claims["session"] = CreateSessKey(username)
	claims["rights"] = userDetails
	claims["exp"] = fmt.Sprintf("%d-%02d-%02d %02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	claims["username"] = userDetails.Username

	sessionIDstr, _ := json.Marshal(sessionIDs)
	var user Users
	if variables.Cache {
		// log session id to redis
		database.RdbCon.Set(ctx, username, string(sessionIDstr), 0)
	} else {
		fmt.Printf("\tusername = %v \t session_id = %v\n", username, string(sessionIDstr))
		err := setSessionID(username, string(sessionIDstr))
		if err != nil {
			log.Println("failed to save sessionID error =", err)
			return "", user, err
		}
	}

	// fmt.Printf("\n\nCreate expiry date = %v\n", claims["exp"])
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", user, err
	}

	return tokenString, userDetails, nil
}

// ValidateJWT validates whether token is valid
func ValidateJWT(tokenStr string) (Users, bool) {
	var user Users
	token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Println("Unexpected signing method: %v", "HMAC")
			return nil, fmt.Errorf("Unexpected signing method: %v", "HMAC")
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return mySigningKey, nil
	})

	// fmt.Println("\n\t token =", token.Claims.(jwt.MapClaims))
	fmt.Println("\n\t token is valid=", token.Valid)

	// check if token is valid and get claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		claimsJson, _ := json.Marshal(claims)
		json.Unmarshal(claimsJson, &user)

		// get today's date to compare with token's expiry
		today := time.Now()

		// get expiry date to compare
		expiry, _ := time.Parse("2006-01-02 15:04", fmt.Sprintf("%v", claims["exp"]))

		// check if token is expired
		if today.After(expiry) {
			fmt.Println("\t token expired")
			return user, false
		}

		fmt.Println("\n\t user =", user.Username)

		return user, true
	}

	// fmt.Println("user =", claims["username"])
	return user, false
}

func AuthenticateRequest(token string) (map[string]interface{}, Users) {
	respMap := make(map[string]interface{})

	details, authentic := ValidateJWT(token)
	if !authentic {
		respMap["response"] = "error"
		respMap["message"] = "authentication error"

		return respMap, details
	}

	return respMap, details
}

// GetRightsInToken returns a users struct containing the user rights
func GetRightsInToken(rights interface{}) (Users, error) {
	var user Users

	// convert interface to json
	jsonStr, err := json.Marshal(rights)
	if err != nil {
		return user, err
	}

	// get json string into users
	err = json.Unmarshal(jsonStr, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// getSessionID fetches sessionID
func getSessionID(username string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sql := `SELECT 
				session_id::varchar
			FROM users 
			WHERE username = $1`

	rows, err := database.PgPool.Query(ctx, sql, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessionIDstr := ""
	for rows.Next() {
		rows.Scan(&sessionIDstr)
	}

	var sessionID []string
	json.Unmarshal([]byte(sessionIDstr), &sessionID)
	return sessionID, nil
}

// setSessionID fetches sessionID
func setSessionID(username, sessID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sql := `UPDATE users SET session_id = $2 WHERE username = $1`

	_, err := database.PgPool.Exec(ctx, sql, username, sessID)
	if err != nil {
		return err
	}

	return nil
}

// FetchSalesTellers returns a list of users with rights to approve sales
func FetchSalesTellers(branch, username string) ([]map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	c := strings.Replace(branch, "'", "|| chr(39) ||", -1)
	br_con := " AND branch = " + c

	if branch == "all" || branch == "All" {
		br_con = " "
	}

	// prepare sql statement
	sql := `SELECT username FROM users 
			WHERE make_sales = True` + br_con

	rows, err := database.PgPool.Query(ctx, sql)
	if err != nil {
		log.Println("error = ", err)
		return nil, err
	}
	defer rows.Close()

	var users []map[string]string

	users = nil
	for rows.Next() {
		var userName string
		entry := make(map[string]string)

		rows.Scan(&userName)

		entry["value"] = userName
		entry["label"] = userName

		users = append(users, entry)
	}

	return users, nil
}

// FetchSalesApprover returns a list of users with rights to approve sales
func FetchSalesApprover(branch, username string) ([]map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// prepare sql statement
	sql := `SELECT username FROM users 
			WHERE approve_sales = True AND branch = $1 AND username != 'JTELLER'`

	rows, err := database.PgPool.Query(ctx, sql, branch)
	if err != nil {
		log.Println("error = ", err)
		return nil, err
	}
	defer rows.Close()

	var users []map[string]string

	users = nil
	for rows.Next() {
		var userName string
		entry := make(map[string]string)

		rows.Scan(&userName)

		entry["value"] = userName
		entry["label"] = userName

		users = append(users, entry)
	}

	return users, nil
}

// FetchCashApprover returns a list of users with rights to approve sales
func FetchCashApprover(branch, username string) ([]map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// prepare sql statement
	sql := `SELECT username FROM users 
			WHERE cash_rollups = True AND branch = $1 AND username != 'JTELLER'`

	rows, err := database.PgPool.Query(ctx, sql, branch)
	if err != nil {
		log.Println("error = ", err)
		return nil, err
	}
	defer rows.Close()

	var users []map[string]string

	for rows.Next() {
		var userName string
		entry := make(map[string]string)

		rows.Scan(&userName)

		entry["value"] = userName
		entry["label"] = userName

		users = append(users, entry)
	}

	return users, nil
}

// FetchReturnApprover returns a list of users with rights to approve sales returns
func FetchReturnApprover(branch, username string) ([]map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// prepare sql statement
	sql := `SELECT username FROM users 
			WHERE approve_sales_returns = True AND branch = $1 AND username != 'JTELLER'`

	rows, err := database.PgPool.Query(ctx, sql, branch)
	if err != nil {
		log.Println("error = ", err)
		return nil, err
	}
	defer rows.Close()

	var users []map[string]string

	for rows.Next() {
		var userName string
		entry := make(map[string]string)

		rows.Scan(&userName)

		entry["value"] = userName
		entry["label"] = userName

		users = append(users, entry)
	}

	return users, nil
}

// FetchCreditApprover returns a list of users with rights to approve sales
func FetchCreditApprover(branch, username string) ([]map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// prepare sql statement
	sql := `SELECT 
				username 
			FROM users 
			WHERE approve_credit_sales = True 
				AND username != 'JTELLER'
				AND branch = $1 `

	rows, err := database.PgPool.Query(ctx, sql, branch)
	if err != nil {
		log.Println("error = ", err)
		return nil, err
	}
	defer rows.Close()

	var users []map[string]string

	for rows.Next() {
		var userName string
		entry := make(map[string]string)

		rows.Scan(&userName)

		entry["value"] = userName
		entry["label"] = userName

		users = append(users, entry)
	}

	return users, nil
}

// CreateSessKey generates user's unique session key
func CreateSessKey(username string) string {

	// generate random 8 digit number
	ind0 := random.Intn(9)
	ind1 := random.Intn(9)
	ind2 := random.Intn(9)
	ind3 := random.Intn(9)
	ind4 := random.Intn(9)
	ind5 := random.Intn(9)
	ind6 := random.Intn(9)
	ind7 := random.Intn(9)
	ind8 := random.Intn(9)
	ind9 := random.Intn(9)
	ind10 := random.Intn(9)
	ind11 := random.Intn(9)
	ind12 := random.Intn(9)
	ind13 := random.Intn(9)
	ind14 := random.Intn(9)
	ind15 := random.Intn(9)

	token := fmt.Sprintf("%v%v%v%v%v%v%v%v%v%v%v%v%v%v%v%v%v", username, ind0, ind1, ind2, ind3, ind4, ind5, ind6, ind7, ind8, ind9, ind10, ind11, ind12, ind13, ind14, ind15)
	// fmt.Println("signature token = ", token)

	return token
}

func genApproverToken(username string, validity int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// generate random 8 digit number
	ind1 := random.Intn(9)
	ind2 := random.Intn(9)
	ind3 := random.Intn(9)
	ind4 := random.Intn(9)
	ind5 := random.Intn(9)
	ind6 := random.Intn(9)
	ind7 := random.Intn(9)
	ind8 := random.Intn(9)

	token := fmt.Sprintf("%v%v%v%v%v%v%v%v", ind1, ind2, ind3, ind4, ind5, ind6, ind7, ind8)

	expiry := time.Now().Local().Add(time.Hour * 24 * 14)

	sql := `UPDATE users SET token = $1, token_date = $2 WHERE username = $3`
	_, err := database.PgPool.Exec(ctx, sql, token, expiry)
	if err != nil {
		return err
	}

	return nil
}
