package gdrive

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

var GDriveSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync with Google Drive",
	Long:  `Sync with Google Drive`,
	Run:   GDriveSyncCommand,
}

func init() {
	GDriveSyncAddCmd.Flags().StringP("local", "l", "", "Local path to sync from")
	GDriveSyncAddCmd.Flags().StringP("remote", "g", "", "GDrive path to sync to")

	GDriveSyncCmd.AddCommand(GDriveSyncAddCmd)
}

func GDriveSyncCommand(cmd *cobra.Command, args []string) {
	client, ctx := GetGDriveClient()

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	db := NewDB(DBOptions{}).db // get connection to db
	tx := db.Begin()            // start transaction to update db after iteration
	{
		rows, err := db.Model(&Syncable{}).Rows()
		if err != nil {
			log.Fatalln(err)
		}
		defer rows.Close()

		for rows.Next() {
			var syncable Syncable
			db.ScanRows(rows, &syncable)

			processSyncable(tx, srv, &syncable)
		}
	}
	tx.Commit()
}

func processSyncable(tx *gorm.DB, srv *drive.Service, syncable *Syncable) {
	log.Printf("Syncing %s to %s\n", syncable.LocalPath, syncable.GDrivePath)

	if syncable.GDriveFileId == "" && syncable.IsFolder {
		// create folders to store files in
		log.Println("Creating folder")
		driveCreateFolder(srv, syncable)

		tx.Save(&syncable)
	}

	if syncable.IsFolder {
		// iterate over all files in the folder
		files, err := ioutil.ReadDir(syncable.LocalPath)
		if err != nil {
			log.Fatalln(err)
		}
		for _, file := range files {
			if file.IsDir() {
				// dont recurse into folders
				continue
			}

			// upload file to drive
			log.Printf("Syncing %s\n", file.Name())

			// check if file exists in drive
			// if it does, check if it's the same
			// if it's not... TODO:
		}
	} else {
		// upload file to drive
		// TODO:
	}
}

func driveUploadFile(srv *drive.Service, src string, dest string) {

}

func driveCreateFolder(srv *drive.Service, syncable *Syncable) {
	// get drive path
	gdpath := syncable.GDrivePath

	// drive path can be nested
	// split the path and create folders
	// if they don't exist
	folders := strings.Split(gdpath, "/")

	parent := "root"
	for i, folder := range folders {
		if folder == "" {
			// skip empty folders
			continue
		}

		q := "mimeType='application/vnd.google-apps.folder' and name='" + folder + "'"
		if i > 0 {
			q += " and '" + parent + "' in parents"
		}

		response, err := srv.Files.List().Q(q).Do()
		if err != nil {
			log.Fatalln(err)
		}

		if len(response.Files) > 0 {
			// folder exists
			parent = response.Files[0].Id
			continue
		}

		// folder doesn't exist, create it
		file := &drive.File{
			Name:     folder,
			Parents:  []string{parent},
			MimeType: "application/vnd.google-apps.folder",
		}
		res, err := srv.Files.Create(file).Do()
		if err != nil {
			log.Fatalln(err)
		}
		parent = res.Id
	}

	// update the file id
	syncable.GDriveFileId = parent
}
