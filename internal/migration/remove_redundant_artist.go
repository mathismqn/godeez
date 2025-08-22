package migration

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.etcd.io/bbolt"
)

// RemoveRedundantArtistMigration removes redundant artist names from filenames
// This migration should run after the restructure_directories migration
type RemoveRedundantArtistMigration struct{}

func (m *RemoveRedundantArtistMigration) ID() int {
	return 2
}

func (m *RemoveRedundantArtistMigration) Name() string {
	return "remove_redundant_artist"
}

func (m *RemoveRedundantArtistMigration) Description() string {
	return "Remove redundant artist names from track filenames since artist is already in directory structure"
}

func (m *RemoveRedundantArtistMigration) Run(db *bbolt.DB, configDir string) error {
	log.Printf("Starting remove redundant artist migration...")

	// Get the download directory from the config
	outputDir, err := getOutputDirFromConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to get output directory: %w", err)
	}
	if outputDir == "" {
		return fmt.Errorf("could not determine output directory")
	}

	log.Printf("Migration will operate on directory: %s", outputDir)

	// Find all music files in the tree structure
	var filesToProcess []string
	err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".mp3") || strings.HasSuffix(path, ".flac")) {
			filesToProcess = append(filesToProcess, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	log.Printf("Found %d files to process", len(filesToProcess))

	// Process each file
	for i, filePath := range filesToProcess {
		log.Printf("Processing file %d/%d: %s", i+1, len(filesToProcess), filepath.Base(filePath))

		newPath, needsRename := generateNewFilename(filePath)
		if !needsRename {
			continue
		}

		// Rename the file
		if err := os.Rename(filePath, newPath); err != nil {
			log.Printf("Warning: failed to rename %s to %s: %v", filePath, newPath, err)
			continue
		}

		// Update database record if it exists
		if err := updateDatabasePath(db, filePath, newPath); err != nil {
			log.Printf("Warning: failed to update database for %s: %v", filePath, err)
		}
	}

	log.Printf("Migration completed: processed %d files", len(filesToProcess))
	return nil
}

// generateNewFilename removes redundant artist name from filename
// Returns the new path and whether a rename is needed
func generateNewFilename(filePath string) (string, bool) {
	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)
	ext := filepath.Ext(filename)
	baseFilename := strings.TrimSuffix(filename, ext)

	// Extract artist name from directory structure
	// Path should be: outputDir/Artist/Album/filename
	pathParts := strings.Split(dir, string(os.PathSeparator))
	if len(pathParts) < 2 {
		return filePath, false // Not in expected Artist/Album structure
	}

	artistName := pathParts[len(pathParts)-2] // Artist is second-to-last directory

	// Check if filename contains redundant artist name
	// Pattern: "TrackNumber. Artist - Title" -> "TrackNumber. Title"
	pattern := regexp.MustCompile(`^(\d+\. )(.+?) - (.+)$`)
	matches := pattern.FindStringSubmatch(baseFilename)
	
	if len(matches) != 4 {
		return filePath, false // Doesn't match expected pattern
	}

	trackNumber := matches[1]
	artistInFilename := matches[2]
	title := matches[3]

	// Check if artist in filename matches directory artist
	if strings.EqualFold(artistInFilename, artistName) {
		// Remove redundant artist name
		newFilename := fmt.Sprintf("%s%s%s", trackNumber, title, ext)
		newPath := filepath.Join(dir, newFilename)
		return newPath, true
	}

	return filePath, false
}

// updateDatabasePath updates the database record with the new file path
func updateDatabasePath(db *bbolt.DB, oldPath, newPath string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("downloads"))
		if bucket == nil {
			return nil // Bucket doesn't exist, nothing to update
		}

		// Find the record with the old path
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if string(v) == oldPath {
				// Update with new path
				return bucket.Put(k, []byte(newPath))
			}
		}

		return nil // Record not found, which is okay
	})
}
