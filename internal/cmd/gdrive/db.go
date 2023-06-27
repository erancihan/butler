package gdrive

import (
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Syncable struct {
	gorm.Model
	LocalPath    string
	GDrivePath   string
	GDriveFileId string
	IsFolder     bool
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

	if !options.ShouldMigrate {
		// Migrate the schema
		db.AutoMigrate(&Syncable{})
	}

	return &DB{
		db:   db,
		path: dbpath,
	}
}
