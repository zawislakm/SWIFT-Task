package parser

import (
	"SWIFT-Remitly/internal/database"
	"SWIFT-Remitly/internal/models"
	"encoding/csv"
	"fmt"
	"github.com/jszwec/csvutil"
	"io"
	"log"
	"os"
	"strings"
)

var correctHeaders = []string{
	"COUNTRY ISO2 CODE",
	"SWIFT CODE",
	"CODE TYPE",
	"NAME",
	"ADDRESS",
	"TOWN NAME",
	"COUNTRY NAME",
	"TIME ZONE",
}

func validateHeaders(headers []string) bool {
	if len(headers) != len(correctHeaders) {
		return false
	}
	for i, header := range headers {
		if !strings.EqualFold(header, correctHeaders[i]) {
			return false
		}
	}

	return true
}

func validateFile(file *os.File) (*csvutil.Decoder, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	if fileInfo.Size() == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	csvReader := csv.NewReader(file)

	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	if !validateHeaders(headers) {
		return nil, fmt.Errorf("CSV headers do not match expected headers")
	}

	// Reset the file pointer to the beginning of the file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	decoder, err := csvutil.NewDecoder(csv.NewReader(file))
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV decoder: %w", err)
	}
	return decoder, nil
}

// ParseCSV reads a CSV file and adds the bank data to the database.
// It returns an error if the CSV file cannot be read
// Reads a CSV file line by line, logs if there is an error in decoding the line
// Logs if there is an error in adding the bank to the database
func ParseCSV(db database.Service, csvDataPath string) error {
	log.Println(fmt.Sprintf("Started parsing CSV data from file: %s", csvDataPath))

	file, err := os.OpenFile(csvDataPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close CSV file: %v", err)
		}
	}()

	decoder, err := validateFile(file)
	if err != nil {
		return fmt.Errorf("during CSV validation got: %w", err)
	}

	// Read and process each line
	for {
		var bank models.CreateBankRequest
		if err := decoder.Decode(&bank); err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("failed to decode CSV line: %v", err)
			continue
		}
		if err := db.AddBankFromRequest(bank); err != nil {
			log.Printf("failed to add bank from request: %v", err)
		}
	}

	log.Println("Parsing finished")
	return nil
}
