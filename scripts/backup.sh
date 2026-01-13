#!/bin/bash
set -e

BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p "$BACKUP_DIR/volumes"
mkdir -p "$BACKUP_DIR/sql"

echo "[$DATE] Starting backup..."

# Volume backup
echo "[$DATE] Creating volume snapshots..."
docker compose stop postgres api

docker run --rm -v openexposuremanagement_postgres-data:/data \
  -v $(pwd)/"$BACKUP_DIR/volumes":/backup \
  alpine tar czf /backup/postgres-$DATE.tar.gz -C /data .

docker compose start postgres api

# SQL dump
echo "[$DATE] Creating SQL dump..."
docker exec oem-postgres pg_dump -U oem -h localhost \
  -F c -f /tmp/oem.dump oem

docker cp oem-postgres:/tmp/oem.dump \
  "$BACKUP_DIR/sql/oem-$DATE.dump"

docker exec oem-postgres rm /tmp/oem.dump

echo "[$DATE] Backup complete: $BACKUP_DIR"
echo ""
echo "Volume backups:"
ls -lh "$BACKUP_DIR/volumes"
echo ""
echo "SQL dumps:"
ls -lh "$BACKUP_DIR/sql"
