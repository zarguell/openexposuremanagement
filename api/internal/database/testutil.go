package database

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection and runs migrations
// Call this in your tests to get a clean database
func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	// Get test database URL from env or use default
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://oem:oem@localhost:5432/oem_test?sslmode=disable"
	}

	// Connect to test database
	db, err := sqlx.Connect("postgres", testDBURL)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil
	}

	// Run migrations
	if err := runMigrations(t, db, testDBURL); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Cleanup after test
	t.Cleanup(func() {
		cleanupTestDB(t, db)
	})

	return db
}

// runMigrations runs database migrations on the test database
func runMigrations(t *testing.T, db *sqlx.DB, dbURL string) error {
	t.Helper()

	// Check if migrations have already been run by checking for schema_migrations table or tenants table
	var tableExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&tableExists)
	if err == nil && tableExists {
		// Migrations already applied, skip
		return nil
	}

	// Check if tenants table exists (migrations might have been run manually)
	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'tenants')").Scan(&tableExists)
	if err == nil && tableExists {
		// Tables already exist, skip migrations
		return nil
	}

	// Create database driver instance
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Get migrations path
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		// Default to ../../db/migrations relative to this file
		migrationsPath = "file://../../db/migrations"
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run up migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// cleanupTestDB truncates all tables in the test database
func cleanupTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()

	// List of tables to truncate (in dependency order)
	tables := []string{
		"finding_instances",
		"finding_definition_aliases",
		"finding_definitions",
		"asset_identifiers",
		"assets",
		"api_keys",
		"user_roles",
		"roles",
		"users",
		"tenants",
		"intel_sync_runs",
		"intel_cve",
	}

	// Truncate each table with CASCADE to handle foreign keys
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}

	// Close database connection
	db.Close()
}

// CreateTestTenant creates a test tenant in the database with default roles
func CreateTestTenant(t *testing.T, db *sqlx.DB) int64 {
	t.Helper()

	// Generate unique tenant name using timestamp
	tenantName := "test-tenant-" + time.Now().Format("20060102-150405.000")

	var tenantID int64
	err := db.QueryRow(
		"INSERT INTO tenants (name) VALUES ($1) RETURNING id",
		tenantName,
	).Scan(&tenantID)

	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	// Create default roles for this tenant
	_, err = db.Exec(
		"INSERT INTO roles (tenant_id, name) VALUES ($1, $2), ($1, $3), ($1, $4)",
		tenantID, "admin", "analyst", "viewer",
	)
	if err != nil {
		t.Fatalf("Failed to create default roles: %v", err)
	}

	return tenantID
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, db *sqlx.DB, tenantID int64, email string) int64 {
	t.Helper()

	var userID int64
	err := db.QueryRow(
		`INSERT INTO users (tenant_id, email, display_name, status)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		tenantID, email, "Test User", "active",
	).Scan(&userID)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

// CreateTestAdminUser creates a test admin user with roles
func CreateTestAdminUser(t *testing.T, db *sqlx.DB, tenantID int64) (userID int64, apiKey string) {
	t.Helper()

	// Create user
	userID = CreateTestUser(t, db, tenantID, "admin@example.com")

	// Get admin role
	var adminRoleID int64
	err := db.QueryRow("SELECT id FROM roles WHERE name = 'admin' AND tenant_id = $1", tenantID).Scan(&adminRoleID)
	if err != nil {
		t.Fatalf("Failed to get admin role: %v", err)
	}

	// Assign admin role
	_, err = db.Exec(
		"INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)",
		userID, adminRoleID,
	)
	if err != nil {
		t.Fatalf("Failed to assign admin role: %v", err)
	}

	// Create API key
	apiKey = "test-api-key-admin"
	apiKeyHash := hashAPIKey(t, apiKey)

	_, err = db.Exec(
		`INSERT INTO api_keys (tenant_id, name, key_hash, scopes_json, source_binding)
		VALUES ($1, $2, $3, $4, $5)`,
		tenantID, "test-key", apiKeyHash, `["ingest:vm"]`, nil,
	)
	if err != nil {
		t.Fatalf("Failed to create API key: %v", err)
	}

	return userID, apiKey
}

// hashAPIKey creates a hash of an API key (simplified for testing)
func hashAPIKey(t *testing.T, key string) string {
	t.Helper()
	// For testing, we'll use a simple hash
	// In production, this should use bcrypt or similar
	return key + "-hash"
}
