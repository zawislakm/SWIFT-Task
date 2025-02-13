package parser_test

import (
	"SWIFT-Remitly/internal/models"
	"SWIFT-Remitly/internal/parser"
	"bytes"
	"encoding/csv"
	"log"
	"os"
	"testing"
)

// MockService is a mock implementation of the database.Service interface
type MockService struct{}

func (m *MockService) Close() error {
	return nil
}

func (m *MockService) GetBankBySwiftCode(swiftCode string) (models.Bank, error) {
	return models.Bank{}, nil
}

func (m *MockService) GetBanksByISO2Code(iso2Code string) (models.CountrySWIFTCode, error) {
	return models.CountrySWIFTCode{}, nil
}

func (m *MockService) AddBankFromRequest(requestData models.CreateBankRequest) error {
	return nil
}

func (m *MockService) DeleteBankBySwiftCode(swiftCode string) error {
	return nil
}

var (
	correctHeaders = []string{
		"COUNTRY ISO2 CODE",
		"SWIFT CODE",
		"CODE TYPE",
		"NAME",
		"ADDRESS",
		"TOWN NAME",
		"COUNTRY NAME",
		"TIME ZONE",
	}

	incorrectHeaders = []string{
		"COUNTRY ISO2 CODE",
		"SWIFT CODE",
		"CODE TYPE",
		"NAME",
	}
)

func createMockCSV(headers []string, data [][]string) *os.File {

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)
	if err := writer.Write(headers); err != nil {
		log.Fatalf("failed to write headers: %v", err)
	}
	if err := writer.WriteAll(data); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
	writer.Flush()

	tmpFile, err := os.CreateTemp("", "test*.csv")
	if err != nil {
		log.Fatalf("failed to create temp file: %v", err)
	}

	if _, err := tmpFile.Write(buf.Bytes()); err != nil {
		log.Fatalf("Failed to write to temp file: %v", err)
	}

	return tmpFile
}

type parseCSVTestCase struct {
	name     string
	headers  []string
	data     [][]string
	expected bool
}

func runTestParseCSV(t *testing.T, testCases []parseCSVTestCase) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile := createMockCSV(tc.headers, tc.data)

			defer func() {
				if err := os.Remove(tmpFile.Name()); err != nil {
					log.Printf("Failed to remove temp file: %v", err)
				}
			}()

			if err := tmpFile.Close(); err != nil {
				log.Printf("Failed to close temp file: %v", err)
			}

			err := parser.ParseCSV(&MockService{}, tmpFile.Name())
			if tc.expected && err != nil {
				log.Fatalf("Name: %v, expected nil, got %v", tc.name, err)
			}
			if !tc.expected && err == nil {
				log.Fatalf("Name: %v, expected error, got nil", tc.name)
			}
		})
	}
}

func TestParseCSV(t *testing.T) {

	testCases := []parseCSVTestCase{
		{
			name:    "Valid CSV data",
			headers: correctHeaders,
			data: [][]string{
				{"AL", "AAISALTRXXX", "BIC11", "UNITED BANK OF ALBANIA SH.A", "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023", "TIRANA", "ALBANIA", "Europe/Tirane"},
				{"BG", "ABIEBGS1XXX", "BIC11", "ABV INVESTMENTS LTD", "TSAR ASEN 20  VARNA, VARNA, 9002", "VARNA", "BULGARIA", "Europe/Sofia"},
				{"BG", "ADCRBGS1XXX", "BIC11", "ADAMANT CAPITAL PARTNERS AD", "JAMES BOURCHIER BLVD 76A HILL TOWER SOFIA, SOFIA, 1421", "SOFIA", "BULGARIA", "Europe/Sofia"},
			},
			expected: true,
		},
		{
			name:    "Invalid CSV data",
			headers: correctHeaders,
			data: [][]string{
				{"AL", "AAISALTRXXX", "BIC11", "UNITED BANK OF ALBANIA SH.A", "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023", "TIRANA", "ALBANIA", "Europe/Tirane"},
				{"BG", "ABIEBGS1XXX", "BIC11", "ABV INVESTMENTS LTD", "TSAR ASEN 20  VARNA, VARNA, 9002", "VARNA", "BULGARIA", "Europe/Sofia"},
				{"BG", "ADCRBGS1XXX", "BIC11", "ADAMANT CAPITAL PARTNERS AD", "JAMES BOURCHIER BLVD 76A HILL TOWER SOFIA, SOFIA, 1421", "SOFIA", "BULGARIA"},
				{"PL", "TESTTESTXXX"},
			},
			expected: true,
		},
		{
			name:    "Invalid headers",
			headers: incorrectHeaders,
			data: [][]string{
				{"AL", "AAISALTRXXX", "BIC11", "UNITED BANK OF ALBANIA SH.A", "HYRJA 3 RR. DRITAN HOXHA ND. 11 TIRANA, TIRANA, 1023", "TIRANA", "ALBANIA", "Europe/Tirane"},
				{"BG", "ABIEBGS1XXX", "BIC11", "ABV INVESTMENTS LTD", "TSAR ASEN 20  VARNA, VARNA, 9002", "VARNA", "BULGARIA", "Europe/Sofia"},
				{"BG", "ADCRBGS1XXX", "BIC11", "ADAMANT CAPITAL PARTNERS AD", "JAMES BOURCHIER BLVD 76A HILL TOWER SOFIA, SOFIA, 1421", "SOFIA", "BULGARIA", "Europe/Sofia"},
			},
			expected: false,
		},
	}

	runTestParseCSV(t, testCases)
}
