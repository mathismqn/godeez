package migration

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"go.etcd.io/bbolt"
)

// Migration represents a single migration
type Migration struct {
	ID          int
	Name        string
	Description string
	UpFunc      func(*bbolt.DB, string) error // DB and config directory
}

// MigrationRecord tracks applied migrations in the database
type MigrationRecord struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	AppliedAt   time.Time `json:"applied_at"`
	Description string    `json:"description"`
}

var migrationsBucket = []byte("migrations")

// MigrationRegistry holds all available migrations
type MigrationRegistry struct {
	migrations []Migration
}

// NewMigrationRegistry creates a new migration registry
func NewMigrationRegistry() *MigrationRegistry {
	return &MigrationRegistry{
		migrations: make([]Migration, 0),
	}
}

// Register adds a migration to the registry
func (r *MigrationRegistry) Register(migration Migration) {
	r.migrations = append(r.migrations, migration)
}

// RunMigrations executes all pending migrations
func (r *MigrationRegistry) RunMigrations(db *bbolt.DB, cfgDir string) error {
	// Sort migrations by ID
	sort.Slice(r.migrations, func(i, j int) bool {
		return r.migrations[i].ID < r.migrations[j].ID
	})

	// Get applied migrations
	appliedMigrations, err := r.getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[int]bool)
	for _, applied := range appliedMigrations {
		appliedMap[applied.ID] = true
	}

	// Run pending migrations
	for _, migration := range r.migrations {
		if appliedMap[migration.ID] {
			log.Printf("Migration %d (%s) already applied, skipping", migration.ID, migration.Name)
			continue
		}

		log.Printf("Running migration %d: %s", migration.ID, migration.Name)
		log.Printf("Description: %s", migration.Description)

		if err := migration.UpFunc(db, cfgDir); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", migration.ID, migration.Name, err)
		}

		// Record successful migration
		if err := r.recordMigration(db, migration); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.ID, err)
		}

		log.Printf("Migration %d (%s) completed successfully", migration.ID, migration.Name)
	}

	return nil
}

// getAppliedMigrations retrieves all applied migrations from the database
func (r *MigrationRegistry) getAppliedMigrations(db *bbolt.DB) ([]MigrationRecord, error) {
	var records []MigrationRecord

	err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(migrationsBucket)
		if bucket == nil {
			// No migrations bucket means no migrations have been run
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			var record MigrationRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return err
			}
			records = append(records, record)
			return nil
		})
	})

	return records, err
}

// recordMigration saves a migration record to the database
func (r *MigrationRegistry) recordMigration(db *bbolt.DB, migration Migration) error {
	record := MigrationRecord{
		ID:          migration.ID,
		Name:        migration.Name,
		AppliedAt:   time.Now(),
		Description: migration.Description,
	}

	return db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(migrationsBucket)
		if err != nil {
			return err
		}

		data, err := json.Marshal(record)
		if err != nil {
			return err
		}

		key := []byte(fmt.Sprintf("%04d", migration.ID))
		return bucket.Put(key, data)
	})
}

// GetPendingMigrations returns migrations that haven't been applied
func (r *MigrationRegistry) GetPendingMigrations(db *bbolt.DB) ([]Migration, error) {
	appliedMigrations, err := r.getAppliedMigrations(db)
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[int]bool)
	for _, applied := range appliedMigrations {
		appliedMap[applied.ID] = true
	}

	var pending []Migration
	for _, migration := range r.migrations {
		if !appliedMap[migration.ID] {
			pending = append(pending, migration)
		}
	}

	sort.Slice(pending, func(i, j int) bool {
		return pending[i].ID < pending[j].ID
	})

	return pending, nil
}
