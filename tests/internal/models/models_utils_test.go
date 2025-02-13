package models_test

import (
	"SWIFT-Remitly/internal/models"
	"encoding/json"
	"errors"
	"log"
	"testing"
)

type isHeadquarterBankTestCase struct {
	name     string
	bank     models.Bank
	expected bool
}

func runIsHeadquarterBankTests(t *testing.T, testCases []isHeadquarterBankTestCase) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.bank.IsHeadquarterBank() != tc.expected {
				log.Fatalf("For bank %T, expected %v, got %v", tc.bank, tc.expected, tc.bank.IsHeadquarterBank())
			}
		})
	}
}

func TestIsHeadquarterBank(t *testing.T) {
	testCases := []isHeadquarterBankTestCase{
		{"Headquarter bank", models.Bank{SWIFTCode: "TESTTESTXXX"}, true},
		{"Branch bank", models.Bank{SWIFTCode: "TESTTESTNOT"}, false},
	}
	runIsHeadquarterBankTests(t, testCases)
}

type bankMarshalJSONTestCase struct {
	name     string
	bank     models.Bank
	expected string
}

func runBankMarshalJSONTests(t *testing.T, testCases []bankMarshalJSONTestCase) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.bank.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("MarshalJSON() = %v, want %v", string(result), tt.expected)
			}
		})
	}
}

func TestBankMarshalJSON(t *testing.T) {
	testCases := []bankMarshalJSONTestCase{
		{
			name: "Bank without branches",
			bank: models.Bank{
				Address:   models.BankAddress{Address: "Main St", Town: models.BankTown{Town: "Main Town"}},
				Name:      models.BankName{Name: "Main Bank"},
				Country:   models.BankCountry{ISO2Code: "US", CountryName: "UNITED STATES"},
				SWIFTCode: "TESTTESTNOT",
			},
			expected: `{"address":"Main St","bankName":"Main Bank","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":false,"swiftCode":"TESTTESTNOT"}`,
		},
		{
			name: "Headquarter bank with branches",
			bank: models.Bank{
				Address:   models.BankAddress{Address: "Main St", Town: models.BankTown{Town: "Main Town"}},
				Name:      models.BankName{Name: "Main Bank"},
				Country:   models.BankCountry{ISO2Code: "US", CountryName: "UNITED STATES"},
				SWIFTCode: "TESTTESTXXX",
				Branches: []models.Bank{
					{
						Address:   models.BankAddress{Address: "Branch St", Town: models.BankTown{Town: "Branch Town"}},
						Name:      models.BankName{Name: "Branch 1"},
						Country:   models.BankCountry{ISO2Code: "US", CountryName: "UNITED STATES"},
						SWIFTCode: "TESTTESTNOT",
					},
				},
			},
			expected: `{"address":"Main St","bankName":"Main Bank","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":true,"swiftCode":"TESTTESTXXX","branches":[{"address":"Branch St","bankName":"Branch 1","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":false,"swiftCode":"TESTTESTNOT"}]}`,
		},
		{
			name: "Headquarter bank with branches with extra fields",
			bank: models.Bank{
				Address:   models.BankAddress{Address: "Main St", Town: models.BankTown{Town: "Main Town"}},
				Name:      models.BankName{Name: "Main Bank"},
				Country:   models.BankCountry{ISO2Code: "US", CountryName: "UNITED STATES"},
				SWIFTCode: "TESTTESTXXX",
				TimeZone:  models.TimeZone{TimeZone: "UTC"},
				CodeType:  models.CodeType{CodeType: "TestCodeType"},
				Branches: []models.Bank{
					{
						Address:   models.BankAddress{Address: "Branch St", Town: models.BankTown{Town: "Branch Town"}},
						Name:      models.BankName{Name: "Branch 1"},
						Country:   models.BankCountry{ISO2Code: "US", CountryName: "UNITED STATES"},
						TimeZone:  models.TimeZone{TimeZone: "UTC"},
						CodeType:  models.CodeType{CodeType: "TestCodeType"},
						SWIFTCode: "TESTTESTNOT",
					},
				},
			},
			expected: `{"address":"Main St","bankName":"Main Bank","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":true,"swiftCode":"TESTTESTXXX","branches":[{"address":"Branch St","bankName":"Branch 1","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":false,"swiftCode":"TESTTESTNOT"}]}`,
		},
		{
			name: "Headquarter bank without branches",
			bank: models.Bank{
				Address:   models.BankAddress{Address: "Main St", Town: models.BankTown{Town: "Main Town"}},
				Name:      models.BankName{Name: "Main Bank"},
				Country:   models.BankCountry{ISO2Code: "US", CountryName: "UNITED STATES"},
				SWIFTCode: "TESTTESTXXX",
				Branches:  []models.Bank{},
			},
			expected: `{"address":"Main St","bankName":"Main Bank","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":true,"swiftCode":"TESTTESTXXX","branches":[]}`,
		},
		{
			name: "Headquarter bank without branches with extra fields",
			bank: models.Bank{
				Address:   models.BankAddress{Address: "Main St", Town: models.BankTown{Town: "Main Town"}},
				Name:      models.BankName{Name: "Main Bank"},
				Country:   models.BankCountry{ISO2Code: "US", CountryName: "UNITED STATES"},
				SWIFTCode: "TESTTESTXXX",
				TimeZone:  models.TimeZone{TimeZone: "UTC"},
				CodeType:  models.CodeType{CodeType: "TestCodeType"},
				Branches:  []models.Bank{},
			},
			expected: `{"address":"Main St","bankName":"Main Bank","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":true,"swiftCode":"TESTTESTXXX","branches":[]}`,
		},
		{
			name: "Headquarter bank with empty Country/CountryName without branches",
			bank: models.Bank{
				Address:   models.BankAddress{Address: "Main St", Town: models.BankTown{Town: "Main Town"}},
				Name:      models.BankName{Name: "Main Bank"},
				Country:   models.BankCountry{ISO2Code: "US", CountryName: ""},
				SWIFTCode: "TESTTESTXXX",
				Branches:  []models.Bank{},
			},
			expected: `{"address":"Main St","bankName":"Main Bank","countryISO2":"US","isHeadquarter":true,"swiftCode":"TESTTESTXXX","branches":[]}`,
		},
	}
	runBankMarshalJSONTests(t, testCases)
}

type unmarshalJSONTestCase struct {
	name        string
	jsonData    string
	expected    models.CreateBankRequest
	expectError bool
}

func runUnmarshalJSONTests(t *testing.T, testsCases []unmarshalJSONTestCase) {
	for _, tt := range testsCases {
		t.Run(tt.name, func(t *testing.T) {
			var result models.CreateBankRequest
			err := json.Unmarshal([]byte(tt.jsonData), &result)

			if tt.expectError {
				var errRequestInvalid *models.ErrRequestInvalid
				if !errors.As(err, &errRequestInvalid) {
					t.Errorf("Expected ErrRequestInvalid error, but got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("UnmarshalJSON() = %+v, want %+v", result, tt.expected)
				}
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	testCases := []unmarshalJSONTestCase{
		{
			name: "Valid JSON data",
			jsonData: `{
				"address": "Main St",
				"bankName": "Main Bank",
				"countryISO2": "US",
				"countryName": "UNITED STATES",
				"swiftCode": "TESTTESTXXX",
				"isHeadquarter": true
			}`,
			expected: models.CreateBankRequest{
				Address:       "Main St",
				BankName:      "Main Bank",
				ISO2Code:      "US",
				CountryName:   "UNITED STATES",
				SWIFTCode:     "TESTTESTXXX",
				IsHeadquarter: true,
			},
			expectError: false,
		},
		{
			name: "Wrong isHeadquarter value",
			jsonData: `{
				"address": "Main St",
				"bankName": "Main Bank",
				"countryISO2": "US",
				"countryName": "UNITED STATES",
				"swiftCode": "TESTTESTXXX",
				"isHeadquarter": false
			}`,
			expected:    models.CreateBankRequest{},
			expectError: true,
		},
		{
			name: "Wrong countryISO2 value",
			jsonData: `{
				"address": "Main St",
				"bankName": "Main Bank",
				"countryISO2": "xxx",
				"countryName": "UNITED STATES",
				"swiftCode": "TESTTESTXXX",
				"isHeadquarter": true
			}`,
			expected:    models.CreateBankRequest{},
			expectError: true,
		},
		{
			name: "Wrong countryName value",
			jsonData: `{
				"address": "Main St",
				"bankName": "Main Bank",
				"countryISO2": "xxx",
				"countryName": "united states",
				"swiftCode": "TESTTESTXXX",
				"isHeadquarter": true
			}`,
			expected:    models.CreateBankRequest{},
			expectError: true,
		},
		{
			name: "Wrong swiftCode value",
			jsonData: `{
				"address": "Main St",
				"bankName": "Main Bank",
				"countryISO2": "xxx",
				"countryName": "united states",
				"swiftCode": "testtestxx",
				"isHeadquarter": true
			}`,
			expected:    models.CreateBankRequest{},
			expectError: true,
		},
		{
			name: "Wrong bankName value",
			jsonData: `{
				"address": "Main St",
				"bankName": null,
				"countryISO2": "US",
				"countryName": "UNITED STATES",
				"swiftCode": "TESTTESTXXX",
				"isHeadquarter": true
			}`,
			expected:    models.CreateBankRequest{},
			expectError: true,
		},
	}

	runUnmarshalJSONTests(t, testCases)
}
