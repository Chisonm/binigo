package binigo

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
)

// Migration interface that all migrations must implement
type Migration interface {
	Up(db *sql.DB) error
	Down(db *sql.DB) error
	Name() string
}

// Migrator handles database migrations
type Migrator struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// Register adds a migration to the migrator
func (m *Migrator) Register(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// Run executes all pending migrations
func (m *Migrator) Run() error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Get list of already run migrations
	ran, err := m.getRanMigrations()
	if err != nil {
		return fmt.Errorf("failed to get ran migrations: %v", err)
	}

	// Sort migrations by name (timestamp)
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Name() < m.migrations[j].Name()
	})

	// Run pending migrations
	executed := 0
	for _, migration := range m.migrations {
		if m.hasRun(migration.Name(), ran) {
			continue
		}

		log.Printf("Running migration: %s", migration.Name())

		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("migration %s failed: %v", migration.Name(), err)
		}

		if err := m.recordMigration(migration.Name()); err != nil {
			return fmt.Errorf("failed to record migration %s: %v", migration.Name(), err)
		}

		log.Printf("✅ Migrated: %s", migration.Name())
		executed++
	}

	if executed == 0 {
		log.Println("No pending migrations")
	} else {
		log.Printf("✅ Migrated %d migration(s)", executed)
	}

	return nil
}

// Rollback rolls back the last migration
func (m *Migrator) Rollback() error {
	// Get list of ran migrations
	ran, err := m.getRanMigrations()
	if err != nil {
		return fmt.Errorf("failed to get ran migrations: %v", err)
	}

	if len(ran) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	// Get the last migration
	lastMigration := ran[len(ran)-1]

	// Find the migration
	var migration Migration
	for _, m := range m.migrations {
		if m.Name() == lastMigration {
			migration = m
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found", lastMigration)
	}

	log.Printf("Rolling back: %s", migration.Name())

	if err := migration.Down(m.db); err != nil {
		return fmt.Errorf("rollback %s failed: %v", migration.Name(), err)
	}

	if err := m.deleteMigration(migration.Name()); err != nil {
		return fmt.Errorf("failed to delete migration record %s: %v", migration.Name(), err)
	}

	log.Printf("✅ Rolled back: %s", migration.Name())

	return nil
}

// Reset rolls back all migrations
func (m *Migrator) Reset() error {
	ran, err := m.getRanMigrations()
	if err != nil {
		return err
	}

	// Rollback in reverse order
	for i := len(ran) - 1; i >= 0; i-- {
		if err := m.Rollback(); err != nil {
			return err
		}
	}

	return nil
}

// Status shows the status of all migrations
func (m *Migrator) Status() error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Get list of ran migrations
	ran, err := m.getRanMigrations()
	if err != nil {
		return fmt.Errorf("failed to get ran migrations: %v", err)
	}

	ranMap := make(map[string]bool)
	for _, r := range ran {
		ranMap[r] = true
	}

	// Sort migrations by name (timestamp)
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Name() < m.migrations[j].Name()
	})

	// Print header
	fmt.Println("Migration                                        | Status")
	fmt.Println("------------------------------------------------+--------")

	if len(m.migrations) == 0 {
		fmt.Println("No migrations registered")
		return nil
	}

	// Print each migration
	for _, migration := range m.migrations {
		status := "Pending"
		if ranMap[migration.Name()] {
			status = "✅ Ran"
		}
		fmt.Printf("%-48s | %s\n", migration.Name(), status)
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			migration VARCHAR(255) NOT NULL UNIQUE,
			batch INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := m.db.Exec(query)
	return err
}

// getRanMigrations returns a list of migrations that have been run
func (m *Migrator) getRanMigrations() ([]string, error) {
	rows, err := m.db.Query("SELECT migration FROM migrations ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var migration string
		if err := rows.Scan(&migration); err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// hasRun checks if a migration has already been run
func (m *Migrator) hasRun(name string, ran []string) bool {
	for _, r := range ran {
		if r == name {
			return true
		}
	}
	return false
}

// recordMigration records a migration as having been run
func (m *Migrator) recordMigration(name string) error {
	// Get current batch number
	var batch int
	err := m.db.QueryRow("SELECT COALESCE(MAX(batch), 0) + 1 FROM migrations").Scan(&batch)
	if err != nil {
		batch = 1
	}

	_, err = m.db.Exec("INSERT INTO migrations (migration, batch) VALUES ($1, $2)", name, batch)
	return err
}

// deleteMigration removes a migration record
func (m *Migrator) deleteMigration(name string) error {
	_, err := m.db.Exec("DELETE FROM migrations WHERE migration = $1", name)
	return err
}
