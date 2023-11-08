package files

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/olad5/file-fort/internal/domain"
	"github.com/olad5/file-fort/internal/infra"
	"github.com/olad5/file-fort/internal/services/auth"
	appErrors "github.com/olad5/file-fort/pkg/errors"
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

func (f *FileService) UploadFile(ctx context.Context, file io.Reader, handler *multipart.FileHeader, folderId string) (domain.File, error) {
	jwtClaims, ok := ctx.Value("jwtClaims").(auth.JWTClaims)
	if ok == false {
		return domain.File{}, fmt.Errorf("error parsing JWTClaims")
	}

	userId := jwtClaims.ID
	filename := handler.Filename

	var folderIdInUUID uuid.UUID
	var err error

	if folderId == "" {
		defaultFolder, err := getDefaultFolder(ctx, f, userId)
		if err != nil {
			return domain.File{}, err
		}
		folderIdInUUID = defaultFolder.ID
	} else {
		folderIdInUUID, err = uuid.Parse(folderId)
		if err != nil {
			return domain.File{}, appErrors.ErrInvalidID
		}

		existingFolder, err := f.folderRepo.GetFolderByFolderId(ctx, folderIdInUUID)
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
		FolderId:     folderIdInUUID,
		FileName:     filename,
	}

	err = f.fileRepo.SaveFile(ctx, newFile)
	if err != nil {
		return domain.File{}, err
	}
	return newFile, nil
}

func (f *FileService) DownloadFile(ctx context.Context, fileId uuid.UUID) (string, error) {
	jwtClaims, ok := ctx.Value("jwtClaims").(auth.JWTClaims)
	if ok == false {
		return "", fmt.Errorf("error parsing JWTClaims")
	}

	userId := jwtClaims.ID

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

func (f *FileService) CreateFolder(ctx context.Context, folderName string) (domain.Folder, error) {
	jwtClaims, ok := ctx.Value("jwtClaims").(auth.JWTClaims)
	if ok == false {
		return domain.Folder{}, fmt.Errorf("error parsing JWTClaims")
	}

	userId := jwtClaims.ID
	var err error

	newFolder := domain.Folder{
		ID:         uuid.New(),
		OwnerId:    userId,
		FolderName: folderName,
	}

	err = f.folderRepo.CreateFolder(ctx, newFolder)
	if err != nil {
		return domain.Folder{}, err
	}

	return newFolder, nil
}

func (f *FileService) GetFilesByFolderId(ctx context.Context, folderId uuid.UUID, pageNumber, rowsPerPage int) ([]domain.File, error) {
	jwtClaims, ok := ctx.Value("jwtClaims").(auth.JWTClaims)
	if ok == false {
		return []domain.File{}, fmt.Errorf("error parsing JWTClaims")
	}

	userId := jwtClaims.ID
	existingFolder, err := f.folderRepo.GetFolderByFolderId(ctx, folderId)
	if err != nil {
		return []domain.File{}, err
	}

	if existingFolder.OwnerId != userId {
		return []domain.File{}, infra.ErrUserNotAuthorized
	}

	files, err := f.fileRepo.GetFilesByFolderId(ctx, existingFolder.ID, pageNumber, rowsPerPage)
	if err != nil {
		return []domain.File{}, err
	}

	return files, nil
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

func (f *FileService) MarkFileAsUnsafe(ctx context.Context, fileId uuid.UUID) error {
	file, err := f.fileRepo.GetFileByFileId(ctx, fileId)
	if err != nil {
		return err
	}

	if file.IsUnsafe == true {
		return nil
	}

	err = f.fileStore.DeleteFile(ctx, file.FileStoreKey)
	if err != nil {
		return err
	}

	err = f.fileRepo.MarkFileAsUnsafe(ctx, file)
	if err != nil {
		return err
	}

	return nil
}
