package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	gonanoid "github.com/matoous/go-nanoid"
)

const uploadDir = "./uploads"

func init() {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create uploads directory: %v\n", err)
	}
}

func handleFileUpload(r *http.Request, formFieldName string) (string, error) {
	file, header, err := r.FormFile(formFieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	//gen filename to be unique
	id, err := gonanoid.Generate("companyafcb1230", 6)
	if err != nil {
		return "", err
	}

	//get file ext
	ext := filepath.Ext(header.Filename)
	filename := id + ext
	filepath := filepath.Join(uploadDir, filename)

	//create file
	dst, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return filename, nil
}

func deleteUploadedFile(filename string) error {
	if filename == "" {
		return nil
	}
	filepath := filepath.Join(uploadDir, filename)
	return os.Remove(filepath)
}
