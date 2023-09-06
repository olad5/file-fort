package files

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/olad5/go-cloud-backup-system/internal/domain"
	"github.com/olad5/go-cloud-backup-system/internal/infra"
	appErrors "github.com/olad5/go-cloud-backup-system/pkg/errors"
)

type FileService struct {
	fileRepo   infra.FileRepository
	fileStore  infra.FileStore
	folderRepo infra.FolderRepository
}

func NewFileService(fileRepo infra.FileRepository, folderRepo infra.FolderRepository, fileStore infra.FileStore) (*FileService, error) {
	if fileRepo == nil {
		return &FileService{}, fmt.Errorf("FileService failed to initialize, fileRepo is nil")
	}
	if fileStore == nil {
		return &FileService{}, fmt.Errorf("FileService failed to initialize, fileStore is nil")
	}
	if folderRepo == nil {
		return &FileService{}, fmt.Errorf("FileService failed to initialize, folderRepo is nil")
	}
	return &FileService{fileRepo, fileStore, folderRepo}, nil
}

func (f *FileService) UploadFile(ctx context.Context, file io.Reader, handler *multipart.FileHeader) (domain.File, error) {
	userId, err := uuid.Parse(ctx.Value("userId").(string))
	if err != nil {
		return domain.File{}, fmt.Errorf("error parsing userId to UUID:%w", err)
	}

	filename := handler.Filename

	var folderId uuid.UUID

	if ctx.Value("folderId") == "" {
		defaultFolder, err := getDefaultFolder(ctx, f, userId)
		if err != nil {
			return domain.File{}, err
		}
		folderId = defaultFolder.ID
	} else {
		folderId, err := uuid.Parse(ctx.Value("folderId").(string))
		if err != nil {
			return domain.File{}, appErrors.ErrInvalidID
		}

		existingFolder, err := f.folderRepo.GetFolderByFolderId(ctx, folderId)
		if err != nil {
			return domain.File{}, fmt.Errorf("error getting folder: %w", err)
		}

		if existingFolder.OwnerId != userId {
			return domain.File{}, infra.ErrUserNotAuthorized
		}
	}

	fileStoreKey, err := f.fileStore.SaveToFileStore(ctx, filename, file)
	if err != nil {
		return domain.File{}, fmt.Errorf("unable to save to file Store :%w", err)
	}

	newFile := domain.File{
		ID:           uuid.New(),
		OwnerId:      userId,
		FileStoreKey: fileStoreKey,
		FolderId:     folderId,
		FileName:     filename,
	}

	err = f.fileRepo.SaveFile(ctx, newFile)
	if err != nil {
		return domain.File{}, err
	}
	return newFile, nil
}

func (f *FileService) DownloadFile(ctx context.Context, fileId uuid.UUID) (string, error) {
	userId, err := uuid.Parse(ctx.Value("userId").(string))
	if err != nil {
		return "", fmt.Errorf("error parsing userId to UUID:%w", err)
	}

	file, err := f.fileRepo.GetFileByFileId(ctx, fileId)
	if err != nil {
		return "", err
	}
	if file.OwnerId != userId {
		return "", infra.ErrUserNotAuthorized
	}

	fileUrl, err := f.fileStore.GetDownloadUrl(ctx, file.FileStoreKey)
	return fileUrl, nil
}

func getDefaultFolder(ctx context.Context, f *FileService, userId uuid.UUID) (domain.Folder, error) {
	existingFolder, err := f.folderRepo.GetFolderByFolderId(ctx, userId)
	if err == nil {
		return existingFolder, nil
	}

	const DefaultFolderName = "home"

	newFolder := domain.Folder{
		ID:         userId,
		OwnerId:    userId,
		FolderName: DefaultFolderName,
	}

	err = f.folderRepo.CreateFolder(ctx, newFolder)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("error creating default folder: %w", err)
	}
	return newFolder, nil
}
