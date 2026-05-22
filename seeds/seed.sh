#!/bin/sh

set -e

PGPASSWORD=postgres
export PGPASSWORD

# Wait until a table exists and is queryable
wait_for_table() {
  local host=$1
  local port=$2
  local db=$3
  local table=$4
  echo "⏳  Waiting for table '$table' in $db@$host..."
  until psql -h "$host" -p "$port" -U postgres -d "$db" \
        -c "SELECT 1 FROM $table LIMIT 1;" > /dev/null 2>&1; do
    sleep 2
  done
  echo "✅  Table '$table' is ready."
}

# Wait until a specific column exists in a table.
# This is used to confirm that ALTER TABLE migrations (which run after the
# initial CREATE TABLE) have also completed before seeding begins.
wait_for_column() {
  local host=$1
  local port=$2
  local db=$3
  local table=$4
  local col=$5
  echo "⏳  Waiting for column '$col' on '$table' in $db@$host..."
  until psql -h "$host" -p "$port" -U postgres -d "$db" \
        -c "SELECT $col FROM $table LIMIT 0;" > /dev/null 2>&1; do
    sleep 2
  done
  echo "✅  Column '$col' on '$table' is ready."
}

# --- restaurant_db ---
# Migrations run via docker-entrypoint-initdb.d (all .up.sql files).
# Wait for 'stock' column from migration 005 (last schema change) to confirm
# all five migrations have finished.
wait_for_column postgres-restaurant 5432 restaurant_db menu_items stock

# --- order_db ---
# Migrations run by order-service on startup.
# Migration 005 (last) adds delivery_address / restaurant_id / courier_id to
# orders. The seed INSERT uses those columns, so we must wait for them rather
# than just waiting for the orders table (migration 001) to exist.
wait_for_column postgres-order 5432 order_db orders delivery_address

# --- user_db ---
# Migrations run by user-service on startup.
# Migration 004 adds promo_assigned to users (used in the seed INSERT).
# Migration 005 (last) creates notifications_log. Waiting for that table
# guarantees all five migrations — including the promo_assigned column —
# have completed.
wait_for_table postgres-user 5432 user_db notifications_log

echo "🌱  Seeding restaurant_db..."
psql -h postgres-restaurant -p 5432 -U postgres -d restaurant_db \
     -v ON_ERROR_STOP=1 \
     -f /seeds/restaurant-seed.sql
echo "✅  restaurant_db seeded."

echo "🌱  Seeding order_db..."
psql -h postgres-order -p 5432 -U postgres -d order_db \
     -v ON_ERROR_STOP=1 \
     -f /seeds/order-seed.sql
echo "✅  order_db seeded."

echo "🌱  Seeding user_db..."
psql -h postgres-user -p 5432 -U postgres -d user_db \
     -v ON_ERROR_STOP=1 \
     -f /seeds/user-seed.sql
echo "✅  user_db seeded."

echo ""
echo "🎉  All databases seeded! Demo credentials:"
echo "    Regular user : demo@aitu.kz    / demo1234"
echo "    Student user : student@aitu.kz / demo1234  (promo: AITU2025 → -15%)"
echo "    Courier      : courier@aitu.kz / demo1234"
echo "    Promo codes  : WELCOME10 (-10%)  SUMMER20 (-20%)  FIRSTORDER (-25%)"
