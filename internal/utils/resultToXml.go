package utils

import (
	"corpPR4/internal/models"
	"encoding/xml"
	"fmt"
	"os"
	"path"
	"time"
)

func SaveGameResultToXML(result *models.GameResult) error {
	if result == nil {
		return fmt.Errorf("empty game result")
	}

	data, err := xml.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("xml marshal error: %w", err)
	}

	filename := fmt.Sprintf("game_result_%s.xml", time.Now().Format("20060102_150405"))
	filePath := path.Join("results", filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("file create error: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("xml file closing error: %w", cerr)
		}
	}()

	_, err = file.Write([]byte(xml.Header))
	if err != nil {
		return fmt.Errorf("file write error: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("file write error: %w", err)
	}

	return err
}
