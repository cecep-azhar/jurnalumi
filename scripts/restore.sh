#!/bin/sh
set -e

# Script Uji Restore pg_dump terenkripsi
# Sesuai QA-P1-29 & QA-DECISIONS.md

ENC_FILE="$1"
TARGET_DB=${TARGET_DB:-"jurnalumi_restore_test"}
DB_USER=${POSTGRES_USER:-"postgres"}
DB_HOST=${POSTGRES_HOST:-"127.0.0.1"}
DB_PORT=${POSTGRES_PORT:-"5432"}
ENCRYPTION_KEY=${BACKUP_ENCRYPTION_KEY:-"default-secure-backup-key-jurnalumi-2026"}

if [ -z "$ENC_FILE" ]; then
    echo "Usage: $0 <path-to-encrypted-backup.enc> [target_db]"
    exit 1
fi

if [ ! -f "$ENC_FILE" ]; then
    echo "Error: File $ENC_FILE not found."
    exit 1
fi

TEMP_SQL_GZ="/tmp/restore_temp_$$.sql.gz"

echo "[$(date)] Decrypting backup file..."
openssl enc -d -aes-256-cbc -pbkdf2 -in "$ENC_FILE" -out "$TEMP_SQL_GZ" -pass pass:"$ENCRYPTION_KEY"

echo "[$(date)] Preparing database ${TARGET_DB}..."
PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "DROP DATABASE IF EXISTS ${TARGET_DB};"
PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE ${TARGET_DB};"

echo "[$(date)] Restoring data into ${TARGET_DB}..."
gunzip -c "$TEMP_SQL_GZ" | PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$TARGET_DB"

rm -f "$TEMP_SQL_GZ"
echo "[$(date)] Restore verified successfully on database: ${TARGET_DB}"
