package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pa-pe/wedyta/model"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (s *Service) CheckUploadPermission(ctx *gin.Context, req model.UploadCheckRequest) (model.UploadCheckResponse, error) {
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

	// check for model presence
	mConfig := s.loadModelConfig(ctx, req.Model, nil)
	if mConfig == nil {
		return model.UploadCheckResponse{}, errors.New("incomplete model=" + req.Model)
	}

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

func (s *Service) ProcessImageUpload(ctx *gin.Context) (string, error) {
	if !s.UploadsConfigured {
		return "", errors.New("wedyta uploads not configured")
	}

	recordID := ctx.PostForm("record_id")
	modelName := ctx.PostForm("model")
	field := ctx.PostForm("field")

	if recordID == "" || modelName == "" || field == "" {
		return "", errors.New("missing required parameters")
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return "", errors.New("no file uploaded")
	}

	if !isAllowedExtension(file.Filename) {
		return "", errors.New("invalid file type")
	}

	originalName := sanitizeFileName(file.Filename)

	// check for model presence
	mConfig := s.loadModelConfig(ctx, modelName, nil)
	if mConfig == nil {
		return "", errors.New("incomplete model=" + modelName)
	}

	if !isValidUploadPathComponent(modelName) || !isValidUploadPathComponent(recordID) {
		return "", errors.New("invalid model name or record ID")
	}

	// path to: uploads/modelName/recordID/
	uploadDir := filepath.Join("uploads", modelName, recordID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Absolute path to file
	savePath := filepath.Join(uploadDir, originalName)

	// Check: if the file already exists - error
	if _, err := os.Stat(savePath); err == nil {
		return "", errors.New("file with this name already exists")
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("unable to check file: %w", err)
	}

	// validate MimeType
	ok, err := isAllowedImageMimeType(file)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("invalid file content: not an image")
	}

	// Save
	if err := saveUploadedFile(ctx, file, savePath); err != nil {
		return "", errors.New("failed to save file")
	}

	// Relative path to image
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

// isValidUploadPathComponent checks if the path component is valid
func isValidUploadPathComponent(input string) bool {
	valid := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return valid.MatchString(input)
}

// isAllowedImageMimeType checks that the file is indeed an image
func isAllowedImageMimeType(fileHeader *multipart.FileHeader) (bool, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return false, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return false, fmt.Errorf("unable to read file: %w", err)
	}

	mimeType := http.DetectContentType(buffer[:n])
	if strings.HasPrefix(mimeType, "image/") {
		return true, nil
	}
	return false, nil
}

func saveUploadedFile(c *gin.Context, file *multipart.FileHeader, path string) error {
	return c.SaveUploadedFile(file, path)
}
