package gdrive

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var GDriveSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync with Google Drive",
	Long:  `Sync with Google Drive`,
}

var GDriveSyncAddCmd = &cobra.Command{
	Use:   "add [flags] (path/to/local/) [path/to/gdrive]",
	Short: "Add a new syncable path",
	Long:  `Add a new syncable path`,
	Run:   GDriveSyncAddCommand,
}

func init() {
	GDriveSyncAddCmd.Flags().StringP("local", "l", "", "Local path to sync from")
	GDriveSyncAddCmd.Flags().StringP("remote", "g", "", "GDrive path to sync to")

	GDriveSyncCmd.AddCommand(GDriveSyncAddCmd)
}

func GDriveSyncAddCommand(cmd *cobra.Command, args []string) {
	// get local path
	var localPath string

	localPath, err := cmd.Flags().GetString("local")
	if err != nil {
		log.Fatalln(err)
	}
	if len(args) > 0 { // override local path if provided as arg, favor arg over flag
		localPath = args[0]
	}
	if localPath == "" {
		log.Fatalln("You must at least provide local path")
	}
	localPath = filepath.Clean(localPath)
	// check if local path exists
	fstat, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		log.Fatalln("Local path does not exist")
	}
	if err != nil {
		log.Fatalln(err)
	}

	// get gdrive path if exists
	gdrivePath, err := cmd.Flags().GetString("remote")
	if err != nil {
		log.Fatalln(err)
	}
	if len(args) > 1 { // override gdrive path if provided as arg, favor arg over flag
		gdrivePath = args[1]
	}

	// check if gdrive path is provided, otherwise set from local path
	if gdrivePath == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalln(err)
		}

		gdrivePath = filepath.Join("/butler", filepath.Clean(strings.Replace(localPath, homedir, "", 1)))
	}

	// Update record in DB.
	db := NewDB(DBOptions{}).db // get connection to db

	syncable := Syncable{
		LocalPath:  localPath,
		GDrivePath: gdrivePath,
		IsFolder:   fstat.IsDir(),
	}

	entry := Syncable{}
	tx := db.Limit(1).Find(&entry, &Syncable{LocalPath: localPath})
	if tx.Error != nil {
		log.Fatalln(tx.Error)
	}
	if tx.RowsAffected > 0 {
		// to update, we need to set the ID, otherwise it will create a new entry
		syncable.ID = entry.ID
	}
	tx.Save(&syncable)

	fmt.Printf("Added '%s' to sync at '%s'\n", localPath, gdrivePath)
}
