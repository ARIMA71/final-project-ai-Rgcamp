package service

import (
	repository "a21hc3NpZ25tZW50/repository/fileRepository"
	"encoding/csv"
	"fmt"
	"strings"
)

type FileService struct {
	Repo *repository.FileRepository
}

func (s *FileService) ProcessFile(fileContent string) (map[string][]string, error) {
	reader := csv.NewReader(strings.NewReader(fileContent))

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	headers := records[0]
	data := make(map[string][]string)

	for _, header := range headers {
		data[header] = []string{}
	}

	for _, row := range records[1:] {
		for i, val := range row {
			data[headers[i]] = append(data[headers[i]], val)
		}
	}

	// fmt.Println(data)
	return data, nil

}
