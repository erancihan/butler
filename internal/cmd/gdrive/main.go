package gdrive

import (
	"github.com/spf13/cobra"
)

var GDriveCmd = &cobra.Command{
	Use:   "gdrive",
	Short: "",
	Long:  "",
}

func init() {
	// Here you will define your flags and configuration settings.
	GDriveCmd.AddCommand(GDriveAuthCmd)
	GDriveCmd.AddCommand(GDriveSyncCmd)
}
