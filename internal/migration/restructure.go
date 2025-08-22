package migration

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/flytam/filenamify"
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
	
	// Register the remove redundant artist migration
	redundantArtistMigration := &RemoveRedundantArtistMigration{}
	registry.Register(Migration{
		ID:          redundantArtistMigration.ID(),
		Name:        redundantArtistMigration.Name(),
		Description: redundantArtistMigration.Description(),
		UpFunc:      redundantArtistMigration.Run,
	})
}

// migrateToTreeStructure handles the migration from old structure to new tree structure
func migrateToTreeStructure(db *bbolt.DB, cfgDir string) error {
	log.Println("Starting directory restructure migration...")

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
		log.Printf("Moving file %d/%d: %s", i+1, len(filesToMove), filepath.Base(moveOp.OldPath))

		if err := moveFile(moveOp.OldPath, moveOp.NewPath); err != nil {
			log.Printf("Warning: Failed to move %s to %s: %v", moveOp.OldPath, moveOp.NewPath, err)
			continue
		}

		successfulMoves++

		// Update database record
		if err := updateDatabaseRecord(db, moveOp.SongID, moveOp.NewPath); err != nil {
			log.Printf("Warning: Failed to update database record for %s: %v", moveOp.SongID, err)
		}
	}

	log.Printf("Migration completed: %d/%d files moved successfully", successfulMoves, len(filesToMove))

	// Clean up empty directories
	if err := cleanupEmptyDirectories(outputDir); err != nil {
		log.Printf("Warning: Failed to clean up empty directories: %v", err)
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

// generateNewPath creates the new tree-structured path from the old path
func generateNewPath(oldPath, outputDir string) (string, error) {
	fileName := filepath.Base(oldPath)
	
	// Parse filename to extract artist and track info
	// Expected formats:
	// "01. Artist - Track.mp3" (from albums)
	// "Artist - Track.mp3" (from singles/playlists)
	
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)
	
	// Remove track number if present
	trackNumRemoved := nameWithoutExt
	if strings.Contains(nameWithoutExt, ". ") {
		parts := strings.SplitN(nameWithoutExt, ". ", 2)
		if len(parts) == 2 {
			trackNumRemoved = parts[1]
		}
	}
	
	// Split artist and track
	if !strings.Contains(trackNumRemoved, " - ") {
		return "", fmt.Errorf("cannot parse artist and track from filename: %s", fileName)
	}
	
	parts := strings.SplitN(trackNumRemoved, " - ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid filename format: %s", fileName)
	}
	
	artist := strings.TrimSpace(parts[0])
	// Track name is embedded in the filename, so we don't need to extract it separately
	
	// Determine album from the parent directory or use "Unknown Album"
	album := "Unknown Album"
	oldDir := filepath.Dir(oldPath)
	
	// If the parent directory is not the base output directory, use it as album
	if oldDir != outputDir && filepath.Base(oldDir) != "GoDeez" {
		parentDirName := filepath.Base(oldDir)
		
		// Check if parent directory looks like "Artist - Album" format
		if strings.Contains(parentDirName, " - ") {
			albumParts := strings.SplitN(parentDirName, " - ", 2)
			if len(albumParts) == 2 {
				album = strings.TrimSpace(albumParts[1])
			}
		} else if parentDirName != "Singles" && parentDirName != "Playlists" {
			// Use the directory name as album if it's not a known single/playlist folder
			album = parentDirName
		}
	}
	
	// Sanitize path components
	artist, _ = filenamify.Filenamify(artist, filenamify.Options{})
	album, _ = filenamify.Filenamify(album, filenamify.Options{})
	fileName, _ = filenamify.Filenamify(fileName, filenamify.Options{})
	
	return filepath.Join(outputDir, artist, album, fileName), nil
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
