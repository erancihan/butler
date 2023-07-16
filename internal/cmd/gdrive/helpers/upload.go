package gdrive_helpers

import (
	"errors"
	"fmt"

	"google.golang.org/api/drive/v3"
)

type UploadOptions struct {
	Name       string // The name of the file to upload
	LocalPath  string // The path to the file to upload
	UploadPath string // The path to the folder to upload to
	FolderId   string // The parent folder ID to upload to
	// MimeType   string // The MIME type of the file to upload

	Service *drive.Service // The service to use to upload the file
}

type UploadResult struct {
	FolderId string // The parent folder ID of the uploaded file
	FileId   string // The file ID of the uploaded file
}

func Upload(opts *UploadOptions) (*UploadResult, error) {
	if opts.Service == nil {
		return nil, errors.New("Service is nil")
	}

	fmt.Printf("Uploading %s to %s\n", opts.LocalPath, opts.UploadPath)

	if opts.FolderId == "" {
		// check if file path exists in gdrive
		fid, err := Mkdirs(opts.Service, opts.UploadPath)
		if err != nil {
			return nil, err
		}
		opts.FolderId = fid
	}

	return &UploadResult{
		FolderId: opts.FolderId,
	}, nil
}
