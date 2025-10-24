package cmd

import (
	"fmt"

	"github.com/api-monolith-template/internal/infrastructure"
	"github.com/api-monolith-template/internal/seeder"
	"github.com/api-monolith-template/internal/util"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database with initial data",
	Run: func(cmd *cobra.Command, args []string) {
		// Init DB connection
		infrastructure.InitializeDBConn()
		sqlDB, err := infrastructure.DB.DB()
		util.ContinueOrFatal(err)

		err = sqlDB.Ping()
		util.ContinueOrFatal(err)

		if err := seeder.Run(infrastructure.DB); err != nil {
			util.ContinueOrFatal(err)
		}
		fmt.Println("✅ Seeder completed.")
	},
}

var seedFlushCmd = &cobra.Command{
	Use:   "seed-flush",
	Short: "Flush seeded data from the database",
	Run: func(cmd *cobra.Command, args []string) {
		infrastructure.InitializeDBConn()
		sqlDB, err := infrastructure.DB.DB()
		util.ContinueOrFatal(err)

		err = sqlDB.Ping()
		util.ContinueOrFatal(err)

		if err := seeder.Flush(infrastructure.DB); err != nil {
			fmt.Println("❌ Gagal flush seed:", err)
		} else {
			fmt.Println("✅ Semua seed berhasil dihapus.")
		}
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
	rootCmd.AddCommand(seedFlushCmd)
}
