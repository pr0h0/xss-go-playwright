package services

import (
	"bytes"
	"encoding/json"
)

type JsonService struct {
	enc  *json.Encoder
	buff bytes.Buffer
}

var jsonServiceInstance *JsonService = nil

// Singleton instance of JsonService
func GetJsonService() (*JsonService, error) {
	if jsonServiceInstance == nil {
		jsonServiceInstance = &JsonService{
			buff: bytes.Buffer{},
		}
		encoder := json.NewEncoder(&jsonServiceInstance.buff)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		jsonServiceInstance.enc = encoder
	}

	return jsonServiceInstance, nil
}

func (js *JsonService) GetBuffer() []byte {
	value := js.buff.Bytes()
	js.buff.Reset()

	return value
}

// Convert struct to JSON
func (js *JsonService) StructToJson(data interface{}) (string, error) {
	err := js.enc.Encode(data)
	if err != nil {
		return "", err
	}
	return string(js.GetBuffer()), nil
}

// Convert JSON to struct
func (js *JsonService) JsonToStruct(jsonData string, data interface{}) error {
	err := json.Unmarshal([]byte(jsonData), data)
	if err != nil {
		return err
	}
	return nil
}

// Convert JSON to map
func (js *JsonService) JsonToMap(jsonData string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Convert map to JSON
func (js *JsonService) MapToJson(data map[string]interface{}) (string, error) {
	err := js.enc.Encode(data)
	if err != nil {
		return "", err
	}
	return string(js.GetBuffer()), nil
}

// Convert Array to string
func (js *JsonService) ArrayToString(data []string) (string, error) {
	err := js.enc.Encode(data)
	if err != nil {
		return "", err
	}
	return string(js.GetBuffer()), nil
}

// Convert string to Array
func (js *JsonService) StringToArray(jsonData string) ([]string, error) {
	var data []string
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
