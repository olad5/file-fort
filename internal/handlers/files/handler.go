package handlers

import (
	"errors"

	"github.com/olad5/file-fort/internal/usecases/files"
)

type FileHandler struct {
	fileService files.FileService
}

func NewFileHandler(fileService files.FileService) (*FileHandler, error) {
	if fileService == (files.FileService{}) {
		return nil, errors.New("file service cannot be empty")
	}

	return &FileHandler{fileService}, nil
}
