package migration

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/flytam/filenamify"
	"github.com/mathismqn/godeez/internal/tags"
	"go.etcd.io/bbolt"
)

// RegisterAllMigrations registers all available migrations
func RegisterAllMigrations(registry *MigrationRegistry) {
	registry.Register(Migration{
		ID:          1,
		Name:        "restructure_directories",
		Description: "Migrate from flat/artist-album structure to Artist/Album/Track tree structure",
		UpFunc:      migrateToTreeStructure,
	})
}

// migrateToTreeStructure handles the migration from old structure to new tree structure
func migrateToTreeStructure(db *bbolt.DB, cfgDir string, dryRun bool) error {
	log.Println("Starting directory restructure migration...")
	prefix := ""
	if dryRun {
		prefix = "[DRY RUN] "
		log.Println("DRY RUN: Showing what would be done without making changes")
	}

	// Get output directory from config
	outputDir, err := getOutputDirFromConfig(cfgDir)
	if err != nil {
		return fmt.Errorf("failed to get output directory: %w", err)
	}

	if outputDir == "" {
		outputDir = filepath.Join(os.Getenv("HOME"), "Music", "GoDeez")
	}

	log.Printf("Migration will operate on directory: %s", outputDir)

	// Check if directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		log.Println("Output directory doesn't exist, migration not needed")
		return nil
	}

	// Collect all download records that need path updates
	var recordsToUpdate []DownloadRecord
	var filesToMove []FileMoveOperation

	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("tracks"))
		if bucket == nil {
			log.Println("No tracks bucket found, migration not needed")
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			var info DownloadRecord
			if err := json.Unmarshal(v, &info); err != nil {
				log.Printf("Warning: Failed to unmarshal record for key %s: %v", string(k), err)
				return nil // Continue with other records
			}

			// Check if file exists at current path
			if _, err := os.Stat(info.Path); os.IsNotExist(err) {
				log.Printf("Warning: File not found at %s, skipping", info.Path)
				return nil // Continue with other records
			}

			// Generate new path based on file structure
			newPath, err := generateNewPath(info.Path, outputDir)
			if err != nil {
				log.Printf("Warning: Failed to generate new path for %s: %v", info.Path, err)
				return nil // Continue with other records
			}

			// Only migrate if the path actually changed
			if newPath != info.Path {
				recordsToUpdate = append(recordsToUpdate, info)
				filesToMove = append(filesToMove, FileMoveOperation{
					OldPath: info.Path,
					NewPath: newPath,
					SongID:  info.SongID,
				})
			}

			return nil
		})
	})

	if err != nil {
		return fmt.Errorf("failed to scan existing records: %w", err)
	}

	log.Printf("Found %d files to migrate", len(filesToMove))

	if len(filesToMove) == 0 {
		log.Println("No files need migration")
		return nil
	}

	// Perform file moves
	successfulMoves := 0
	for i, moveOp := range filesToMove {
		log.Printf(prefix+"Moving file %d/%d:", i+1, len(filesToMove))
		log.Printf("    FROM: %s", moveOp.OldPath)
		log.Printf("    TO:   %s", moveOp.NewPath)

		if dryRun {
			successfulMoves++
			continue
		}

		if err := moveFile(moveOp.OldPath, moveOp.NewPath); err != nil {
			log.Printf("Warning: Failed to move file: %v", err)
			continue
		}

		successfulMoves++

		// Update database record
		if err := updateDatabaseRecord(db, moveOp.SongID, moveOp.NewPath); err != nil {
			log.Printf("Warning: Failed to update database record for %s: %v", moveOp.SongID, err)
		} else {
			log.Printf("Database record updated for song %s", moveOp.SongID)
		}
	}

	log.Printf(prefix+"Migration completed: %d/%d files moved successfully", successfulMoves, len(filesToMove))
	if !dryRun {

		// Clean up empty directories
		if err := cleanupEmptyDirectories(outputDir); err != nil {
			log.Printf("Warning: Failed to clean up empty directories: %v", err)
		}
	}

	return nil
}

type DownloadRecord struct {
	SongID     string `json:"song_id"`
	Quality    string `json:"quality"`
	Path       string `json:"path"`
	Hash       string `json:"hash"`
	Downloaded string `json:"downloaded_at"`
}

type FileMoveOperation struct {
	OldPath string
	NewPath string
	SongID  string
}

// generateNewPath creates the new tree-structured path from the old path using ID3 tags
func generateNewPath(oldPath, outputDir string) (string, error) {
	// Read metadata from the audio file
	metadata, err := tags.ReadMetadata(oldPath)
	if err != nil {
		return "", fmt.Errorf("failed to read metadata from %s: %w", oldPath, err)
	}

	// Extract required fields
	artist := strings.TrimSpace(metadata.AlbumArtist)
	if artist == "" {
		artist = strings.TrimSpace(metadata.Artist)
	}
	if artist == "" {
		return "", fmt.Errorf("no artist information found in metadata for %s", oldPath)
	}

	album := strings.TrimSpace(metadata.Album)
	if album == "" {
		album = "Unknown Album"
	}

	title := strings.TrimSpace(metadata.Title)
	if title == "" {
		// Fall back to filename without extension if no title in metadata
		fileName := filepath.Base(oldPath)
		title = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	}

	// Format the filename with track number if available
	fileName := title
	if metadata.TrackNumber != "" {
		fileName = fmt.Sprintf("%s. %s", metadata.TrackNumber, title)
	}

	// Add file extension
	ext := filepath.Ext(oldPath)
	fileName += ext

	// Sanitize path components for filesystem safety
	artist, _ = filenamify.Filenamify(artist, filenamify.Options{})
	album, _ = filenamify.Filenamify(album, filenamify.Options{})
	fileName, _ = filenamify.Filenamify(fileName, filenamify.Options{})

	// Build the new path: outputDir/Artist/Album/Track.ext
	newPath := filepath.Join(outputDir, artist, album, fileName)

	return newPath, nil
}

// moveFile moves a file from oldPath to newPath, creating directories as needed
func moveFile(oldPath, newPath string) error {
	// Create target directory
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", newDir, err)
	}

	// Check if target file already exists
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("target file already exists: %s", newPath)
	}

	// Move the file
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// updateDatabaseRecord updates the path in the database record
func updateDatabaseRecord(db *bbolt.DB, songID, newPath string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("tracks"))
		if bucket == nil {
			return fmt.Errorf("tracks bucket not found")
		}

		// Get existing record
		data := bucket.Get([]byte(songID))
		if data == nil {
			return fmt.Errorf("record not found for song ID %s", songID)
		}

		var record DownloadRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return err
		}

		// Update path
		record.Path = newPath

		// Save updated record
		updatedData, err := json.Marshal(record)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(songID), updatedData)
	})
}

// cleanupEmptyDirectories removes empty directories after migration
func cleanupEmptyDirectories(outputDir string) error {
	return filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue despite errors
		}

		if !info.IsDir() || path == outputDir {
			return nil
		}

		// Check if directory is empty
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil // Continue despite errors
		}

		if len(entries) == 0 {
			log.Printf("Removing empty directory: %s", path)
			os.Remove(path) // Ignore errors
		}

		return nil
	})
}

// getOutputDirFromConfig reads the output directory from config file
func getOutputDirFromConfig(cfgDir string) (string, error) {
	configPath := filepath.Join(cfgDir, "config.toml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", nil // Config doesn't exist, will use default
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	// Simple TOML parsing for output_dir
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "output_dir") && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, "\"'")
				if value != "" {
					return value, nil
				}
			}
		}
	}

	return "", nil // No output_dir specified in config
}
