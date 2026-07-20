-- ============================================================
-- Speed Sales — Login Service Migration
-- Safe to run multiple times (fully idempotent).
-- ============================================================

-- ── Users table ───────────────────────────────────────────
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

    -- operational flags
    remote_login        BOOL NOT NULL DEFAULT FALSE,
    adopt_stockcount    BOOL NOT NULL DEFAULT FALSE,
    complete_stockcount BOOL NOT NULL DEFAULT FALSE,
    pos_settings        BOOL NOT NULL DEFAULT FALSE,

    -- dispatch
    post_dispatch          BOOL NOT NULL DEFAULT FALSE,
    approve_dispatch       BOOL NOT NULL DEFAULT FALSE,
    grant_post_dispatch    BOOL NOT NULL DEFAULT FALSE,
    grant_approve_dispatch BOOL NOT NULL DEFAULT FALSE,

    -- receiving
    post_receive          BOOL NOT NULL DEFAULT FALSE,
    approve_receive       BOOL NOT NULL DEFAULT FALSE,
    grant_post_receive    BOOL NOT NULL DEFAULT FALSE,
    grant_approve_receive BOOL NOT NULL DEFAULT FALSE,

    -- orders
    post_orders          BOOL NOT NULL DEFAULT FALSE,
    approve_orders       BOOL NOT NULL DEFAULT FALSE,
    grant_post_orders    BOOL NOT NULL DEFAULT FALSE,
    grant_approve_orders BOOL NOT NULL DEFAULT FALSE,

    -- reports
    access_sales_reports       BOOL NOT NULL DEFAULT FALSE,
    grant_access_sales_reports BOOL NOT NULL DEFAULT FALSE,

    -- sales / POS
    make_sales          BOOL NOT NULL DEFAULT FALSE,
    approve_sales       BOOL NOT NULL DEFAULT FALSE,
    accept_payment      BOOL NOT NULL DEFAULT FALSE,
    grant_make_sales    BOOL NOT NULL DEFAULT FALSE,
    grant_approve_sales BOOL NOT NULL DEFAULT FALSE,
    grant_accept_payment BOOL NOT NULL DEFAULT FALSE,

    -- cash
    cash_office              BOOL NOT NULL DEFAULT FALSE,
    cash_rollups             BOOL NOT NULL DEFAULT FALSE,
    approve_cash_rollups     BOOL NOT NULL DEFAULT FALSE,
    grant_cash_office        BOOL NOT NULL DEFAULT FALSE,
    grant_cash_rollups       BOOL NOT NULL DEFAULT FALSE,
    grant_approve_cash_rollups BOOL NOT NULL DEFAULT FALSE,

    -- returns
    sales_returns             BOOL NOT NULL DEFAULT FALSE,
    approve_sales_returns     BOOL NOT NULL DEFAULT FALSE,
    grant_sales_returns       BOOL NOT NULL DEFAULT FALSE,
    grant_approve_sales_returns BOOL NOT NULL DEFAULT FALSE,

    -- laybyes / credit
    laybyes                  BOOL NOT NULL DEFAULT FALSE,
    approve_credit_sales     BOOL NOT NULL DEFAULT FALSE,
    grant_laybyes            BOOL NOT NULL DEFAULT FALSE,
    grant_approve_credit_sales BOOL NOT NULL DEFAULT FALSE,

    -- mpesa
    activate_mpesa       BOOL NOT NULL DEFAULT FALSE,
    grant_activate_mpesa BOOL NOT NULL DEFAULT FALSE,

    -- cheques
    post_cheques          BOOL NOT NULL DEFAULT FALSE,
    approve_cheques       BOOL NOT NULL DEFAULT FALSE,
    grant_post_cheques    BOOL NOT NULL DEFAULT FALSE,
    grant_approve_cheques BOOL NOT NULL DEFAULT FALSE,

    -- users management
    create_users      BOOL NOT NULL DEFAULT FALSE,
    delete_users      BOOL NOT NULL DEFAULT FALSE,
    grant_create_users BOOL NOT NULL DEFAULT FALSE,
    grant_delete_users BOOL NOT NULL DEFAULT FALSE,

    -- invoices / pricing
    price_change        BOOL NOT NULL DEFAULT FALSE,
    ammend_invoice      BOOL NOT NULL DEFAULT FALSE,
    grant_price_change  BOOL NOT NULL DEFAULT FALSE,
    grant_ammend_invoice BOOL NOT NULL DEFAULT FALSE,

    -- stock
    create_stock       BOOL NOT NULL DEFAULT FALSE,
    link_stock         BOOL NOT NULL DEFAULT FALSE,
    audit_stock        BOOL NOT NULL DEFAULT FALSE,
    complete_stock_take BOOL NOT NULL DEFAULT FALSE,
    grant_create_stock BOOL NOT NULL DEFAULT FALSE,
    grant_audit_stock  BOOL NOT NULL DEFAULT FALSE,

    -- accounts
    accounts             BOOL NOT NULL DEFAULT FALSE,
    aprove_accounts      BOOL NOT NULL DEFAULT FALSE,
    grant_accounts       BOOL NOT NULL DEFAULT FALSE,
    grant_aprove_accounts BOOL NOT NULL DEFAULT FALSE,

    -- ledger
    ledger            BOOL NOT NULL DEFAULT FALSE,
    recon_ledger      BOOL NOT NULL DEFAULT FALSE,
    grant_ledger      BOOL NOT NULL DEFAULT FALSE,
    grant_recon_ledger BOOL NOT NULL DEFAULT FALSE,

    -- smart card
    create_scard      BOOL NOT NULL DEFAULT FALSE,
    grant_create_scard BOOL NOT NULL DEFAULT FALSE,

    -- production
    produce       BOOL NOT NULL DEFAULT FALSE,
    grant_produce BOOL NOT NULL DEFAULT FALSE,

    CONSTRAINT users_username_key PRIMARY KEY (username)
);

-- ── Column patches (for existing deployments) ──────────────
ALTER TABLE users ADD COLUMN IF NOT EXISTS accept_payment        BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_accept_payment  BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS delete_users          BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_delete_users    BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS adopt_stockcount      BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS complete_stockcount   BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS remote_login          BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS cash_office           BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_cash_office     BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS activate_mpesa        BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_activate_mpesa  BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS post_cheques          BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS approve_cheques       BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_post_cheques    BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_approve_cheques BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS ammend_invoice        BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_ammend_invoice  BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS produce               BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS grant_produce         BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile               JSONB NOT NULL DEFAULT '{}';

-- ── Default Admin user (insert if username does not exist) ──
INSERT INTO users (
    username,   first_name,     user_class, password,  reset,
    -- operational
    remote_login, adopt_stockcount, complete_stockcount, pos_settings,
    -- dispatch
    post_dispatch, approve_dispatch, grant_post_dispatch, grant_approve_dispatch,
    -- receiving
    post_receive, approve_receive, grant_post_receive, grant_approve_receive,
    -- orders
    post_orders, approve_orders, grant_post_orders, grant_approve_orders,
    -- reports
    access_sales_reports, grant_access_sales_reports,
    -- sales
    make_sales, approve_sales, accept_payment,
    grant_make_sales, grant_approve_sales, grant_accept_payment,
    -- cash
    cash_office, cash_rollups, approve_cash_rollups,
    grant_cash_office, grant_cash_rollups, grant_approve_cash_rollups,
    -- returns
    sales_returns, approve_sales_returns,
    grant_sales_returns, grant_approve_sales_returns,
    -- laybyes
    laybyes, approve_credit_sales,
    grant_laybyes, grant_approve_credit_sales,
    -- mpesa
    activate_mpesa, grant_activate_mpesa,
    -- cheques
    post_cheques, approve_cheques, grant_post_cheques, grant_approve_cheques,
    -- users
    create_users, delete_users, grant_create_users, grant_delete_users,
    -- pricing
    price_change, ammend_invoice, grant_price_change, grant_ammend_invoice,
    -- stock
    create_stock, link_stock, audit_stock, complete_stock_take,
    grant_create_stock, grant_audit_stock,
    -- accounts
    accounts, aprove_accounts, grant_accounts, grant_aprove_accounts,
    -- ledger
    ledger, recon_ledger, grant_ledger, grant_recon_ledger,
    -- smart card
    create_scard, grant_create_scard,
    -- production
    produce, grant_produce
)
SELECT
    'Admin', 'Administrator', 'super user', '123', TRUE,
    -- operational
    TRUE, TRUE, TRUE, TRUE,
    -- dispatch
    TRUE, TRUE, TRUE, TRUE,
    -- receiving
    TRUE, TRUE, TRUE, TRUE,
    -- orders
    TRUE, TRUE, TRUE, TRUE,
    -- reports
    TRUE, TRUE,
    -- sales
    TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE,
    -- cash
    TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE,
    -- returns
    TRUE, TRUE,
    TRUE, TRUE,
    -- laybyes
    TRUE, TRUE,
    TRUE, TRUE,
    -- mpesa
    TRUE, TRUE,
    -- cheques
    TRUE, TRUE, TRUE, TRUE,
    -- users
    TRUE, TRUE, TRUE, TRUE,
    -- pricing
    TRUE, TRUE, TRUE, TRUE,
    -- stock
    TRUE, TRUE, TRUE, TRUE,
    TRUE, TRUE,
    -- accounts
    TRUE, TRUE, TRUE, TRUE,
    -- ledger
    TRUE, TRUE, TRUE, TRUE,
    -- smart card
    TRUE, TRUE,
    -- production
    TRUE, TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE username = 'Admin'
);
