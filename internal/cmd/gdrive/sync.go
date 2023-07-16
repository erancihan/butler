package gdrive

import (
	gdrive_helpers "butler/internal/cmd/gdrive/helpers"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

var GDriveSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync with Google Drive",
	Long:  `Sync with Google Drive`,
	Run:   newCommand(),
}

type command struct {
	validate bool
}

func init() {
	GDriveSyncAddCmd.Flags().StringP("local", "l", "", "Local path to sync from")
	GDriveSyncAddCmd.Flags().StringP("remote", "g", "", "GDrive path to sync to")

	GDriveSyncCmd.AddCommand(GDriveSyncAddCmd)
	GDriveSyncCmd.Flags().Bool("validate", false, "Validate syncable paths")
}

func newCommand() func(cmd *cobra.Command, args []string) {
	c := &command{}

	return c.GDriveSyncCommand
}

func (c *command) GDriveSyncCommand(cmd *cobra.Command, args []string) {
	var err error

	c.validate, err = cmd.Flags().GetBool("validate")
	if err != nil {
		log.Fatalln(err)
	}

	client, ctx := GetGDriveClient()

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	db := NewDB(DBOptions{ShouldMigrate: true}).db // get connection to db
	tx := db.Begin()                               // start transaction to update db after iteration
	{
		rows, err := db.Model(&Syncable{}).Rows()
		if err != nil {
			log.Fatalln(err)
		}
		defer rows.Close()

		for rows.Next() {
			var syncable Syncable
			db.ScanRows(rows, &syncable)

			c.processSyncable(tx, srv, &syncable)
		}
	}
	tx.Commit()
}

func (c *command) processSyncable(tx *gorm.DB, srv *drive.Service, syncable *Syncable) {
	log.Printf("Syncing %s to %s\n", syncable.LocalPath, syncable.GDrivePath)

	// first check SyncableItems for changes
	rows, err := tx.Model(&SyncableItem{}).Where("syncable_id = ?", syncable.ID).Rows()
	if err != nil {
		log.Fatalln(err)
	}
	defer rows.Close()

	for rows.Next() {
		var syncableItem SyncableItem
		tx.ScanRows(rows, &syncableItem)

		// check if file still exists
		// TODO: file crud

		fmt.Printf("%+v\n", syncableItem)
	}

	//
	// once we have processed existing entries, we check for new ones
	//
	var files []string
	if syncable.IsFolder {
		files, err = gdrive_helpers.Glob(syncable.LocalPath)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		// if syncable is a file, we should not have arrived here
		// when adding a signle file for syncing, it should have been added to the db
		// as a syncable item.
		// which means we should have arrived at the first for loop, and not this one.
		files = []string{}
	}

	// iterate through all files in the folder
	for _, file := range files {
		excludes := map[string]bool{
			".git":       true,
			".gitignore": true,
			".gitkeep":   true,
			".DS_Store":  true,
		}
		if excludes[filepath.Ext(file)] {
			continue
		}

		stat, err := os.Stat(file)
		if err != nil {
			log.Fatalln(err)
		}
		if stat.IsDir() {
			// skip if entry is a directory
			continue
		}

		// check if file exists in db
		var syncableItem SyncableItem

		tx.Find(&syncableItem, "local_path = ?", file)
		if syncableItem.ID != 0 {
			// file exists
			continue
		}

		// file does not exist, create
		// remove syncable path from file path
		//
		// ensure syncable.LocalPath ends with a slash
		if syncable.LocalPath[len(syncable.LocalPath)-1] != '/' {
			syncable.LocalPath += "/"
		}
		// since this will always be a folder sync, we can assume that
		// syncable's gdrivepath will always be a folder
		if syncable.GDrivePath[len(syncable.GDrivePath)-1] != '/' {
			syncable.GDrivePath += "/"
		}

		syncableItem.FileName = filepath.Base(file)
		syncableItem.LocalPath = file
		syncableItem.GDrivePath = syncable.GDrivePath + file[len(syncable.LocalPath):]
		syncableItem.GDriveFolder = syncableItem.GDrivePath[0 : len(syncableItem.GDrivePath)-len(syncableItem.FileName)]

		// upload file... in a goroutine... TODO
		// upload one by one for now
		//
		// create an upload queue and upload in batches.
		// sort the queue by file size and upload smallest first
		// also have a max number of uploads at a time

		// upload file
		result, err := gdrive_helpers.Upload(&gdrive_helpers.UploadOptions{
			Name:       syncableItem.FileName,
			LocalPath:  syncableItem.LocalPath,
			UploadPath: syncableItem.GDrivePath,
			Service:    srv,
		})
		if err != nil {
			log.Fatalln(err)
		}
		syncableItem.GDriveFolderId = result.FolderId
		syncableItem.GDriveFileId = result.FileId

		tx.Save(&syncableItem)
	}
	tx.Commit()
}
