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
	LocalPath  string
	GDrivePath string
	IsFolder   bool
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
