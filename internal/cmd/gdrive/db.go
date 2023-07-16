package gdrive

import (
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// this is the entry that the user sets
type Syncable struct {
	gorm.Model
	FileName       string
	LocalPath      string
	GDrivePath     string
	GDriveFolder   string
	GDriveFolderId string
	GDriveId       string
	IsFolder       bool
}

// this is the entry that is created by the program will always be a file.
//
// this is the entry that will be kept track of.
type SyncableItem struct {
	gorm.Model

	FileName       string
	LocalPath      string
	GDrivePath     string
	GDriveFolder   string
	GDriveFolderId string
	GDriveFileId   string

	SyncableId int
	Syncable   Syncable
}

type DB struct {
	db   *gorm.DB
	path string
}

type DBOptions struct {
	DbPath        string
	ShouldMigrate bool
}

func NewDB(options DBOptions) *DB {
	dbpath := options.DbPath
	if dbpath == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalln(err)
		}

		dbpath = filepath.Join(homedir, ".config", "butler", "gdrive.db")
	}

	// check if the directory exists
	if _, err := os.Stat(filepath.Dir(dbpath)); os.IsNotExist(err) {
		// create the directory
		err := os.MkdirAll(filepath.Dir(dbpath), 0755)
		if err != nil {
			log.Fatalln(err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
	if err != nil {
		log.Fatalln("failed to connect database")
	}

	if options.ShouldMigrate {
		// Migrate the schema
		db.AutoMigrate(&Syncable{})
		db.AutoMigrate(&SyncableItem{})
	}

	return &DB{
		db:   db,
		path: dbpath,
	}
}
