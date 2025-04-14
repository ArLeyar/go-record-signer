#!/bin/bash

if ! command -v psql &> /dev/null; then
  echo "Error: psql command not found. Please install PostgreSQL client."
  exit 1
fi

DB_URL="postgres://postgres:postgres@localhost:5432/recordsigner"

echo "Using database: ${DB_URL}"

echo -e "\n🔑 SIGNING KEYS"
psql $DB_URL -c "SELECT 
                   count(*) AS total_keys,
                   count(*) FILTER (WHERE in_use = true) AS keys_in_use,
                   count(*) FILTER (WHERE in_use = false) AS keys_available
                 FROM signing_keys;"

echo -e "\n📊 RECORDS BY STATUS"
psql $DB_URL -c "SELECT 
                   status, 
                   count(*) AS count,
                   round(count(*) * 100.0 / NULLIF((SELECT count(*) FROM records), 0), 2) AS percentage
                 FROM records 
                 GROUP BY status 
                 ORDER BY count DESC;"

echo -e "\n✅ Check complete" 