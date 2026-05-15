#!/bin/bash

# Migration Runner for Idempotency Key
# CRITICAL-003 Fix - Add idempotency_key to transactions table
# Date: 2026-05-15

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-simpo}"
DB_USER="${DB_USER:-postgres}"
DRY_RUN=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --host)
            DB_HOST="$2"
            shift 2
            ;;
        --port)
            DB_PORT="$2"
            shift 2
            ;;
        --database)
            DB_NAME="$2"
            shift 2
            ;;
        --user)
            DB_USER="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --dry-run       Show SQL without executing"
            echo "  --host HOST     Database host (default: localhost)"
            echo "  --port PORT     Database port (default: 5432)"
            echo "  --database NAME Database name (default: simpo)"
            echo "  --user USER     Database user (default: postgres)"
            echo "  --help          Show this help"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

MIGRATION_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../migrations"
UP_MIGRATION="$MIGRATION_DIR/20260515120000_add_idempotency_key_to_transactions.up.sql"
DOWN_MIGRATION="$MIGRATION_DIR/20260515120000_add_idempotency_key_to_transactions.down.sql"

echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}Idempotency Key Migration Runner${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""
echo "Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
echo "Migration: $UP_MIGRATION"
echo ""

# Check if migration file exists
if [ ! -f "$UP_MIGRATION" ]; then
    echo -e "${RED}Error: Migration file not found: $UP_MIGRATION${NC}"
    exit 1
fi

# Function to run SQL
run_sql() {
    local sql_file=$1
    local psql_cmd="psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $sql_file"

    if [ "$DRY_RUN" = true ]; then
        echo -e "${YELLOW}DRY RUN - Would execute:${NC}"
        cat "$sql_file"
        echo ""
    else
        echo -e "${GREEN}Executing migration...${NC}"
        if $psql_cmd; then
            echo -e "${GREEN}✓ Migration successful${NC}"
            return 0
        else
            echo -e "${RED}✗ Migration failed${NC}"
            return 1
        fi
    fi
}

# Function to verify migration
verify_migration() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "${YELLOW}DRY RUN - Skipping verification${NC}"
        return 0
    fi

    echo ""
    echo -e "${GREEN}Verifying migration...${NC}"

    # Check if column exists
    local column_check=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
        SELECT COUNT(*)
        FROM information_schema.columns
        WHERE table_name = 'transactions'
          AND column_name = 'idempotency_key';
    ")

    if [ "$column_check" -eq 1 ]; then
        echo -e "${GREEN}✓ Column idempotency_key exists${NC}"
    else
        echo -e "${RED}✗ Column idempotency_key not found${NC}"
        return 1
    fi

    # Check if unique index exists
    local index_check=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
        SELECT COUNT(*)
        FROM pg_indexes
        WHERE tablename = 'transactions'
          AND indexname = 'idx_transactions_idempotency_key';
    ")

    if [ "$index_check" -eq 1 ]; then
        echo -e "${GREEN}✓ Unique index idx_transactions_idempotency_key exists${NC}"
    else
        echo -e "${RED}✗ Unique index not found${NC}"
        return 1
    fi

    # Check for NULL values
    local null_check=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
        SELECT COUNT(*)
        FROM transactions
        WHERE idempotency_key IS NULL;
    ")

    if [ "$null_check" -eq 0 ]; then
        echo -e "${GREEN}✓ All transactions have idempotency_key${NC}"
    else
        echo -e "${YELLOW}⚠ $null_check transactions have NULL idempotency_key${NC}"
    fi

    # Check for legacy format
    local legacy_check=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -tAc "
        SELECT COUNT(*)
        FROM transactions
        WHERE idempotency_key LIKE 'legacy-%';
    ")

    echo -e "${GREEN}✓ $legacy_check legacy transactions have been backfilled${NC}"

    return 0
}

# Main execution
echo -e "${YELLOW}This will add the idempotency_key column to the transactions table.${NC}"
echo ""

if [ "$DRY_RUN" = false ]; then
    read -p "Continue? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        echo "Aborted."
        exit 0
    fi
fi

echo ""
echo "Running migration..."
echo ""

if run_sql "$UP_MIGRATION"; then
    if verify_migration; then
        echo ""
        echo -e "${GREEN}======================================${NC}"
        echo -e "${GREEN}Migration completed successfully!${NC}"
        echo -e "${GREEN}======================================${NC}"
        echo ""
        echo "Next steps:"
        echo "1. Deploy backend code with idempotency support"
        echo "2. Deploy mobile app with UUID generation"
        echo "3. Monitor for unique constraint violations"
        echo "4. Test network retry scenario"
        echo ""
        echo "To rollback: $0 --down"
    else
        echo ""
        echo -e "${RED}Verification failed. Please check the migration.${NC}"
        exit 1
    fi
else
    echo ""
    echo -e "${RED}Migration failed. No changes were applied.${NC}"
    echo ""
    echo "To rollback if partially applied:"
    echo "psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $DOWN_MIGRATION"
    exit 1
fi
