package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) CheckUploadPermission(req model.UploadCheckRequest) (model.UploadCheckResponse, error) {
	if !s.UploadsConfigured {
		return model.UploadCheckResponse{
			Allowed: false,
			Message: "Wedyta uploads not configured.",
		}, nil
	}

	if req.ID == "" {
		return model.UploadCheckResponse{
			Allowed: false,
			Message: "Empty record ID.",
		}, nil
	}

	if req.ID == "0" {
		return model.UploadCheckResponse{
			Allowed: false,
			Message: "Cannot upload while creating a new record.",
		}, nil
	}

	if req.Model == "" || req.Field == "" {
		return model.UploadCheckResponse{}, errors.New("incomplete parameters")
	}

	// TODO: check for module presence

	return model.UploadCheckResponse{
		Allowed: true,
		Message: "",
	}, nil

	//// default false
	//return model.UploadCheckResponse{
	//	Allowed: false,
	//	Message: fmt.Sprintf("Image upload not allowed for model %s, field %s.", req.Model, req.Field),
	//}, nil
}

func (s *Service) ProcessImageUpload(c *gin.Context) (string, error) {
	if !s.UploadsConfigured {
		return "", errors.New("wedyta uploads not configured")
	}

	recordID := c.PostForm("record_id")
	modelName := c.PostForm("model")
	field := c.PostForm("field")

	if recordID == "" || modelName == "" || field == "" {
		return "", errors.New("missing required parameters")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return "", errors.New("no file uploaded")
	}

	if !isAllowedExtension(file.Filename) {
		return "", errors.New("invalid file type")
	}

	originalName := sanitizeFileName(file.Filename)

	// Путь к папке: uploads/modelName/recordID/
	uploadDir := filepath.Join("uploads", modelName, recordID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Абсолютный путь к файлу
	savePath := filepath.Join(uploadDir, originalName)

	// Проверка: если файл уже существует — ошибка
	if _, err := os.Stat(savePath); err == nil {
		return "", errors.New("file with this name already exists")
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("unable to check file: %w", err)
	}

	// Сохраняем
	if err := saveUploadedFile(c, file, savePath); err != nil {
		return "", errors.New("failed to save file")
	}

	// Относительный путь к изображению
	imageURL := "/" + filepath.ToSlash(savePath) // web URL style
	return imageURL, nil
}

func isAllowedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}
	return allowed[ext]
}

func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.", r) {
			return r
		}
		return '_'
	}, name)
	return name
}

func saveUploadedFile(c *gin.Context, file *multipart.FileHeader, path string) error {
	return c.SaveUploadedFile(file, path)
}
