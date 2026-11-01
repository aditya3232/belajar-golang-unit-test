package cmd

import (
	"belajar-golang-unit-test/internal/app"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Menjalankan HTTP API server",
	Long:  "Menjalankan HTTP API server belajar-golang-unit-test",
	Run: func(cmd *cobra.Command, args []string) {
		app.RunServer()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
