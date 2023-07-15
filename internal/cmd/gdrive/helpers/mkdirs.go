package gdrive_helpers

import (
	"strings"

	"google.golang.org/api/drive/v3"
)

func Mkdirs(srv *drive.Service, mkpath string) (string, error) {
	pathv := strings.Split(mkpath, "/")

	parent := "root"
	for i, folder := range pathv {
		if folder == "" {
			// skip empty folders
			continue
		}

		// create query string to search for folder, to see if it exists
		q := "mimeType='application/vnd.google-apps.folder' and name='" + folder + "'"
		if i > 0 {
			// if we're not at the root, search for the parent folder
			q += " and '" + parent + "' in parents"
		}
		// only search for folders that aren't deleted
		q += " and trashed=false"

		response, err := srv.Files.List().Q(q).Do()
		if err != nil {
			return "", err
		}
		if len(response.Files) > 0 {
			// folder exists, set parent to this folder and continue
			parent = response.Files[0].Id
			continue
		}

		// folder doesn't exist, create it
		file := &drive.File{
			Name:     folder,
			Parents:  []string{parent},
			MimeType: "application/vnd.google-apps.folder",
		}
		// use a special color for the *butler* folder
		if folder == "butler" && (i == 0 || i == 1) {
			file.FolderColorRgb = "#4985e7"
		}

		res, err := srv.Files.Create(file).Do()
		if err != nil {
			return "", err
		}

		// set parent to the newly created folder, and continue
		parent = res.Id
	}

	return parent, nil
}
