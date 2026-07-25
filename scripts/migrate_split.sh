#!/usr/bin/env bash
# Login service — split `users` into login/credentials/rights tables
# Usage: ./scripts/migrate_split.sh
# Run AFTER ./scripts/migrate.sh (this reads users.tenant_id, which that
# script is what adds). Run from the Login service root or anywhere; it
# finds its own path.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$SERVICE_DIR/.env"

# ── Colour helpers ─────────────────────────────────────────
GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
ok()   { echo -e "${GREEN}[OK]${NC}  $*"; }
err()  { echo -e "${RED}[ERR]${NC} $*" >&2; exit 1; }
info() { echo -e "${YELLOW}[--]${NC}  $*"; }

# ── Load .env ──────────────────────────────────────────────
if [[ ! -f "$ENV_FILE" ]]; then
  err ".env not found at $ENV_FILE"
fi

set -o allexport
source <(grep -E '^(DB_HOST|DB_PORT|DB_NAME|DB_USER|DB_PASSWORD)\s*=' "$ENV_FILE" \
  | sed 's/[[:space:]]//g; s/="\(.*\)"/=\1/; s/"//g')
set +o allexport

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-logins}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

export PGPASSWORD="$DB_PASSWORD"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Login Service — Split users → login/credentials/rights"
echo "  DB : $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if ! command -v psql &>/dev/null; then
  err "psql not found. Install postgresql-client and retry."
fi

info "Testing database connection..."
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
     -c "SELECT 1" &>/dev/null; then
  err "Cannot connect to $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME — check credentials and that the DB exists."
fi
ok "Connected to $DB_NAME"
echo ""

info "Applying split migration..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
     --set ON_ERROR_STOP=1 \
     -f "$SCRIPT_DIR/split_users_migration.sql"
ok "Migration applied"
echo ""

info "Verifying row counts (login should match users)..."
USERS_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM users;" 2>/dev/null || echo "0")
LOGIN_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM login;")

if [[ "$LOGIN_COUNT" -ge "$USERS_COUNT" ]]; then
  ok "login has $LOGIN_COUNT row(s), users has $USERS_COUNT — backfill complete"
else
  err "login has only $LOGIN_COUNT row(s) but users has $USERS_COUNT — check split_users_migration.sql for errors."
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "  ${GREEN}Migration complete.${NC}"
echo "  NOTE: the legacy 'users' table was left untouched, and"
echo "  Login's Go code still reads/writes it. Cutting the service"
echo "  over to the new tables is a separate follow-up."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
