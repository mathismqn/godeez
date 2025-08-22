package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/mathismqn/godeez/internal/migration"
	"github.com/mathismqn/godeez/internal/store"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database and file structure migrations",
	Long: `Run database and file structure migrations to update GoDeez to the latest format.

This command will:
- Migrate your music files from the old directory structure to the new tree structure
- Update the database records with new file paths
- Clean up empty directories after migration

The migration is safe and will preserve your existing files. If a migration fails, 
the files and database will remain in their current state.

Use --dry-run to preview what changes would be made without actually performing them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		cfgDir := homeDir + "/.godeez"
		configFile := cfgDir + "/config.toml"

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("config file not found at %s. Please run 'godeez download --help' first to create the config", configFile)
		}

		appConfig, err := config.New(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		db, err := store.GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database connection: %w", err)
		}

		registry := migration.NewMigrationRegistry()
		migration.RegisterAllMigrations(registry)

		pending, err := registry.GetPendingMigrations(db)
		if err != nil {
			return fmt.Errorf("failed to check pending migrations: %w", err)
		}

		if len(pending) == 0 {
			fmt.Println("No migrations needed. Your GoDeez installation is up to date!")
			return nil
		}

		fmt.Printf("Found %d pending migration(s):\n\n", len(pending))
		for _, mig := range pending {
			fmt.Printf("  %d. %s\n", mig.ID, mig.Name)
			fmt.Printf("     %s\n\n", mig.Description)
		}

		// In dry run mode, don't ask for confirmation
		if dryRun {
			fmt.Println("Running in DRY RUN mode - no changes will be made to files or database")
			if err := registry.RunMigrations(db, appConfig.ConfigDir, dryRun); err != nil {
				return fmt.Errorf("migration preview failed: %w", err)
			}
			fmt.Println("Dry run completed successfully!")
			return nil
		}

		// Ask for confirmation
		fmt.Print("Do you want to run these migrations? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			return nil
		}

		if err := registry.RunMigrations(db, appConfig.ConfigDir, dryRun); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		fmt.Println("All migrations completed successfully!")
		fmt.Println("\nYour music library has been updated to the new tree structure:")
		fmt.Println("  Artist/Album/Track")
		fmt.Println("\nAny playlists you download in the future will create M3U files")
		fmt.Println("with relative paths to the organized tracks.")

		return nil
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  "Show which migrations have been applied and which are pending.",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		cfgDir := homeDir + "/.godeez"
		configFile := cfgDir + "/config.toml"

		if _, err = os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("config file not found at %s. Please run 'godeez download --help' first to create the config", configFile)
		}

		_, err = config.New(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		db, err := store.GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database connection: %w", err)
		}

		registry := migration.NewMigrationRegistry()
		migration.RegisterAllMigrations(registry)

		pending, err := registry.GetPendingMigrations(db)
		if err != nil {
			return fmt.Errorf("failed to check pending migrations: %w", err)
		}

		if len(pending) == 0 {
			fmt.Println("All migrations have been applied. Your GoDeez installation is up to date!")
		} else {
			fmt.Printf("%d pending migration(s):\n\n", len(pending))
			for _, mig := range pending {
				fmt.Printf("  %d. %s\n", mig.ID, mig.Name)
				fmt.Printf("     %s\n\n", mig.Description)
			}
			fmt.Println("Run 'godeez migrate' to apply these migrations.")
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateStatusCmd)

	migrateCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file (default ~/.godeez/config.toml)")
	migrateCmd.Flags().Bool("dry-run", false, "preview what changes would be made without actually performing them")
}
