package cmd

import (
	"os"

	"butler/internal/cmd/gdrive"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "butler",
	Short: "",
	Long:  "",
}

// Setup adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Setup() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.AddCommand(gdrive.GDriveCmd)
}
