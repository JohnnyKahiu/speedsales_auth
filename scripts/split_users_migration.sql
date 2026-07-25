-- ============================================================
-- Speed Sales — Login Service: split `users` into login /
-- credentials / per-microservice rights tables.
--
-- Run AFTER migrate.sql (this backfill reads users.tenant_id,
-- which migrate.sql is what adds to the legacy table).
--
-- Safe to run multiple times (fully idempotent): every CREATE
-- TABLE uses IF NOT EXISTS and every backfill INSERT skips rows
-- that already exist in the target table. The legacy `users`
-- table is left untouched — this is purely additive. Login's
-- Go code still reads/writes `users` until it's cut over
-- separately; nothing here changes application behaviour yet.
-- ============================================================

-- ── login: pure identity ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS login (
    auto_id    BIGSERIAL PRIMARY KEY,
    username   VARCHAR NOT NULL UNIQUE,
    email      VARCHAR NOT NULL DEFAULT '',
    telephone  VARCHAR NOT NULL DEFAULT '',
    first_name VARCHAR,
    last_name  VARCHAR,
    other_name VARCHAR,
    user_class VARCHAR NOT NULL DEFAULT 'user',
    tenant_id  UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000'
);
CREATE INDEX IF NOT EXISTS idx_login_tenant_id ON login (tenant_id);

-- ── credentials: auth + session/account state, 1:1 with login ──
CREATE TABLE IF NOT EXISTS credentials (
    user_id      BIGINT PRIMARY KEY REFERENCES login (auto_id) ON DELETE CASCADE,
    password     VARCHAR NOT NULL DEFAULT '',
    reset        BOOL NOT NULL DEFAULT FALSE,
    token        VARCHAR(150),
    token_date   TIMESTAMP,
    session_id   VARCHAR NOT NULL DEFAULT '',
    till_num     BIGINT NOT NULL DEFAULT 0,
    till_opened  TIMESTAMP,
    till         BOOL NOT NULL DEFAULT FALSE,
    device       VARCHAR(50),
    branch       VARCHAR NOT NULL DEFAULT '',
    department   VARCHAR NOT NULL DEFAULT '',
    role         VARCHAR NOT NULL DEFAULT '',
    sub_role     VARCHAR,
    access_level VARCHAR,
    stk_location VARCHAR NOT NULL DEFAULT 'shop',
    status       VARCHAR,
    company_id   BIGINT NOT NULL DEFAULT 0,
    remote_login BOOL NOT NULL DEFAULT FALSE,
    profile      JSONB NOT NULL DEFAULT '{}',
    tenant_id    UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000'
);
CREATE INDEX IF NOT EXISTS idx_credentials_tenant_id ON credentials (tenant_id);

-- ── pos_rights: everything the POS service checks ───────────
CREATE TABLE IF NOT EXISTS pos_rights (
    user_id   BIGINT PRIMARY KEY REFERENCES login (auto_id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',

    make_sales bool NOT NULL DEFAULT FALSE, approve_sales bool NOT NULL DEFAULT FALSE, accept_payment bool NOT NULL DEFAULT FALSE,
    grant_make_sales bool NOT NULL DEFAULT FALSE, grant_approve_sales bool NOT NULL DEFAULT FALSE, grant_accept_payment bool NOT NULL DEFAULT FALSE,

    cash_office bool NOT NULL DEFAULT FALSE, cash_rollups bool NOT NULL DEFAULT FALSE, approve_cash_rollups bool NOT NULL DEFAULT FALSE,
    grant_cash_office bool NOT NULL DEFAULT FALSE, grant_cash_rollups bool NOT NULL DEFAULT FALSE, grant_approve_cash_rollups bool NOT NULL DEFAULT FALSE,

    sales_returns bool NOT NULL DEFAULT FALSE, approve_sales_returns bool NOT NULL DEFAULT FALSE,
    grant_sales_returns bool NOT NULL DEFAULT FALSE, grant_approve_sales_returns bool NOT NULL DEFAULT FALSE,

    laybyes bool NOT NULL DEFAULT FALSE, approve_credit_sales bool NOT NULL DEFAULT FALSE,
    grant_laybyes bool NOT NULL DEFAULT FALSE, grant_approve_credit_sales bool NOT NULL DEFAULT FALSE,

    activate_mpesa bool NOT NULL DEFAULT FALSE, grant_activate_mpesa bool NOT NULL DEFAULT FALSE,

    post_cheques bool NOT NULL DEFAULT FALSE, approve_cheques bool NOT NULL DEFAULT FALSE,
    grant_post_cheques bool NOT NULL DEFAULT FALSE, grant_approve_cheques bool NOT NULL DEFAULT FALSE,

    create_scard bool NOT NULL DEFAULT FALSE, grant_create_scard bool NOT NULL DEFAULT FALSE,

    access_sales_reports bool NOT NULL DEFAULT FALSE, grant_access_sales_reports bool NOT NULL DEFAULT FALSE,

    pos_settings bool NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_pos_rights_tenant_id ON pos_rights (tenant_id);

-- ── inventory_rights: everything the Inventory service checks ──
CREATE TABLE IF NOT EXISTS inventory_rights (
    user_id   BIGINT PRIMARY KEY REFERENCES login (auto_id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',

    post_dispatch bool NOT NULL DEFAULT FALSE, approve_dispatch bool NOT NULL DEFAULT FALSE,
    grant_post_dispatch bool NOT NULL DEFAULT FALSE, grant_approve_dispatch bool NOT NULL DEFAULT FALSE,

    post_receive bool NOT NULL DEFAULT FALSE, approve_receive bool NOT NULL DEFAULT FALSE,
    grant_post_receive bool NOT NULL DEFAULT FALSE, grant_approve_receive bool NOT NULL DEFAULT FALSE,

    post_orders bool NOT NULL DEFAULT FALSE, approve_orders bool NOT NULL DEFAULT FALSE,
    grant_post_orders bool NOT NULL DEFAULT FALSE, grant_approve_orders bool NOT NULL DEFAULT FALSE,

    price_change bool NOT NULL DEFAULT FALSE, grant_price_change bool NOT NULL DEFAULT FALSE,
    ammend_invoice bool NOT NULL DEFAULT FALSE, grant_ammend_invoice bool NOT NULL DEFAULT FALSE,

    create_stock bool NOT NULL DEFAULT FALSE, grant_create_stock bool NOT NULL DEFAULT FALSE,
    link_stock bool NOT NULL DEFAULT FALSE, complete_stock_take bool NOT NULL DEFAULT FALSE,
    audit_stock bool NOT NULL DEFAULT FALSE, grant_audit_stock bool NOT NULL DEFAULT FALSE,

    adopt_stockcount bool NOT NULL DEFAULT FALSE, complete_stockcount bool NOT NULL DEFAULT FALSE,
    grant_complete_stockcount bool NOT NULL DEFAULT FALSE,

    produce bool NOT NULL DEFAULT FALSE, grant_produce bool NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_inventory_rights_tenant_id ON inventory_rights (tenant_id);

-- ── accounts_rights: everything the Accounts service checks ────
CREATE TABLE IF NOT EXISTS accounts_rights (
    user_id   BIGINT PRIMARY KEY REFERENCES login (auto_id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',

    accounts bool NOT NULL DEFAULT FALSE, grant_accounts bool NOT NULL DEFAULT FALSE,
    aprove_accounts bool NOT NULL DEFAULT FALSE, grant_aprove_accounts bool NOT NULL DEFAULT FALSE,

    ledger bool NOT NULL DEFAULT FALSE, recon_ledger bool NOT NULL DEFAULT FALSE,
    grant_ledger bool NOT NULL DEFAULT FALSE, grant_recon_ledger bool NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_accounts_rights_tenant_id ON accounts_rights (tenant_id);

-- ── admin_rights: user-management, checked by Login itself ─────
CREATE TABLE IF NOT EXISTS admin_rights (
    user_id   BIGINT PRIMARY KEY REFERENCES login (auto_id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',

    create_users bool NOT NULL DEFAULT FALSE, delete_users bool NOT NULL DEFAULT FALSE,
    grant_create_users bool NOT NULL DEFAULT FALSE, grant_delete_users bool NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_admin_rights_tenant_id ON admin_rights (tenant_id);

-- ============================================================
-- Backfill from the legacy `users` table, if it exists.
-- Idempotent: every INSERT skips rows already present.
-- ============================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users') THEN

        INSERT INTO login (username, email, telephone, first_name, last_name, other_name, user_class, tenant_id)
        SELECT u.username, u.email, u.telephone, u.first_name, u.last_name, u.other_name, u.user_class,
               COALESCE(u.tenant_id, '00000000-0000-0000-0000-000000000000')
        FROM users u
        WHERE NOT EXISTS (SELECT 1 FROM login l WHERE l.username = u.username);

        INSERT INTO credentials (
            user_id, password, reset, token, token_date, session_id, till_num, till_opened, till, device,
            branch, department, role, sub_role, access_level, stk_location, status, company_id, remote_login,
            profile, tenant_id
        )
        SELECT l.auto_id, u.password, u.reset, u.token, u.token_date, u.session_id, u.till_num, u.till_opened, u.till, u.device,
               u.branch, u.department, u.role, u.sub_role, u.access_level, u.stk_location, u.status, u.company_id, u.remote_login,
               u.profile, COALESCE(u.tenant_id, '00000000-0000-0000-0000-000000000000')
        FROM users u
        JOIN login l ON l.username = u.username
        WHERE NOT EXISTS (SELECT 1 FROM credentials c WHERE c.user_id = l.auto_id);

        INSERT INTO pos_rights (
            user_id, tenant_id,
            make_sales, approve_sales, accept_payment, grant_make_sales, grant_approve_sales, grant_accept_payment,
            cash_office, cash_rollups, approve_cash_rollups, grant_cash_office, grant_cash_rollups, grant_approve_cash_rollups,
            sales_returns, approve_sales_returns, grant_sales_returns, grant_approve_sales_returns,
            laybyes, approve_credit_sales, grant_laybyes, grant_approve_credit_sales,
            activate_mpesa, grant_activate_mpesa,
            post_cheques, approve_cheques, grant_post_cheques, grant_approve_cheques,
            create_scard, grant_create_scard,
            access_sales_reports, grant_access_sales_reports,
            pos_settings
        )
        SELECT l.auto_id, COALESCE(u.tenant_id, '00000000-0000-0000-0000-000000000000'),
               u.make_sales, u.approve_sales, u.accept_payment, u.grant_make_sales, u.grant_approve_sales, u.grant_accept_payment,
               u.cash_office, u.cash_rollups, u.approve_cash_rollups, u.grant_cash_office, u.grant_cash_rollups, u.grant_approve_cash_rollups,
               u.sales_returns, u.approve_sales_returns, u.grant_sales_returns, u.grant_approve_sales_returns,
               u.laybyes, u.approve_credit_sales, u.grant_laybyes, u.grant_approve_credit_sales,
               u.activate_mpesa, u.grant_activate_mpesa,
               u.post_cheques, u.approve_cheques, u.grant_post_cheques, u.grant_approve_cheques,
               u.create_scard, u.grant_create_scard,
               u.access_sales_reports, u.grant_access_sales_reports,
               u.pos_settings
        FROM users u
        JOIN login l ON l.username = u.username
        WHERE NOT EXISTS (SELECT 1 FROM pos_rights p WHERE p.user_id = l.auto_id);

        INSERT INTO inventory_rights (
            user_id, tenant_id,
            post_dispatch, approve_dispatch, grant_post_dispatch, grant_approve_dispatch,
            post_receive, approve_receive, grant_post_receive, grant_approve_receive,
            post_orders, approve_orders, grant_post_orders, grant_approve_orders,
            price_change, grant_price_change, ammend_invoice, grant_ammend_invoice,
            create_stock, grant_create_stock, link_stock, complete_stock_take, audit_stock, grant_audit_stock,
            adopt_stockcount, complete_stockcount, grant_complete_stockcount,
            produce, grant_produce
        )
        SELECT l.auto_id, COALESCE(u.tenant_id, '00000000-0000-0000-0000-000000000000'),
               u.post_dispatch, u.approve_dispatch, u.grant_post_dispatch, u.grant_approve_dispatch,
               u.post_receive, u.approve_receive, u.grant_post_receive, u.grant_approve_receive,
               u.post_orders, u.approve_orders, u.grant_post_orders, u.grant_approve_orders,
               u.price_change, u.grant_price_change, u.ammend_invoice, u.grant_ammend_invoice,
               u.create_stock, u.grant_create_stock, u.link_stock, u.complete_stock_take, u.audit_stock, u.grant_audit_stock,
               u.adopt_stockcount, u.complete_stockcount, u.grant_complete_stockcount,
               u.produce, u.grant_produce
        FROM users u
        JOIN login l ON l.username = u.username
        WHERE NOT EXISTS (SELECT 1 FROM inventory_rights i WHERE i.user_id = l.auto_id);

        INSERT INTO accounts_rights (user_id, tenant_id, accounts, grant_accounts, aprove_accounts, grant_aprove_accounts, ledger, recon_ledger, grant_ledger, grant_recon_ledger)
        SELECT l.auto_id, COALESCE(u.tenant_id, '00000000-0000-0000-0000-000000000000'),
               u.accounts, u.grant_accounts, u.aprove_accounts, u.grant_aprove_accounts, u.ledger, u.recon_ledger, u.grant_ledger, u.grant_recon_ledger
        FROM users u
        JOIN login l ON l.username = u.username
        WHERE NOT EXISTS (SELECT 1 FROM accounts_rights a WHERE a.user_id = l.auto_id);

        INSERT INTO admin_rights (user_id, tenant_id, create_users, delete_users, grant_create_users, grant_delete_users)
        SELECT l.auto_id, COALESCE(u.tenant_id, '00000000-0000-0000-0000-000000000000'),
               u.create_users, u.delete_users, u.grant_create_users, u.grant_delete_users
        FROM users u
        JOIN login l ON l.username = u.username
        WHERE NOT EXISTS (SELECT 1 FROM admin_rights ad WHERE ad.user_id = l.auto_id);

    END IF;
END;
$$;
