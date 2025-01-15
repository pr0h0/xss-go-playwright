package services

import (
	"encoding/json"
	"os"
	"xss/utils"
)

type FileService struct {
	jsonService *JsonService
}

var fileServiceInstance *FileService = nil

// Singleton instance of FileService
func GetFileService() (*FileService, error) {
	if fileServiceInstance == nil {
		jsonService, err := GetJsonService()
		utils.HandleErr(err)
		fileServiceInstance = &FileService{
			jsonService: jsonService,
		}
	}

	return fileServiceInstance, nil
}

// Read file as []byte
func (fs *FileService) ReadFileAsBytes(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Read file as string
func (fs *FileService) ReadFileAsString(path string) (string, error) {
	data, err := fs.ReadFileAsBytes(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Read file as JSON
func (fs *FileService) ReadFileAsJSON(path string, jsonStruct interface{}) error {
	data, err := fs.ReadFileAsBytes(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, jsonStruct)
	return err
}

// Write file as []byte
func (fs *FileService) WriteFileAsBytes(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// Write file as string
func (fs *FileService) WriteFileAsString(path string, content string) error {
	return fs.WriteFileAsBytes(path, []byte(content))
}

// Write file as JSON
func (fs *FileService) WriteFileAsJSON(path string, jsonStruct interface{}) error {
	err := fs.jsonService.enc.Encode(jsonStruct)
	if err != nil {
		return err
	}
	return fs.WriteFileAsBytes(path, fs.jsonService.GetBuffer())
}

// Append file as []byte
func (fs *FileService) AppendFileAsBytes(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// Append file as string
func (fs *FileService) AppendFileAsString(path string, content string) error {
	return fs.AppendFileAsBytes(path, []byte(content))
}

// Append file as JSON
func (fs *FileService) AppendFileAsJSON(path string, jsonStruct interface{}) error {
	err := fs.jsonService.enc.Encode(jsonStruct)
	if err != nil {
		return err
	}
	return fs.AppendFileAsBytes(path, fs.jsonService.GetBuffer())
}

// Save file as []byte (overwrite or create new)
func (fs *FileService) SaveFileAsBytes(path string, data []byte) error {
	return fs.WriteFileAsBytes(path, data)
}

// Save file as string (overwrite or create new)
func (fs *FileService) SaveFileAsString(path string, content string) error {
	return fs.SaveFileAsBytes(path, []byte(content))
}

// Save file as JSON (overwrite or create new)
func (fs *FileService) SaveFileAsJSON(path string, jsonStruct interface{}) error {
	err := fs.jsonService.enc.Encode(jsonStruct)
	if err != nil {
		return err
	}
	return fs.SaveFileAsBytes(path, fs.jsonService.GetBuffer())
}
