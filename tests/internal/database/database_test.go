package database

import (
	"SWIFT-Remitly/internal/database"
	"SWIFT-Remitly/internal/models"
	"context"
	"log"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	db, container, err := StartPostgresContainer(context.Background())

	if err != nil {
		panic(err)
	}
	if db == nil || container == nil {
		panic("dbInstance is not properly initialized")
	}

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("could not terminate container: %v", err)
		}
	}()

	var result int
	if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		log.Fatalf("could not ping database: %v", err)
	}
	if result != 1 {
		log.Fatalf("unexpected result after ping: %d", result)
	}

	if err := migrate(db); err != nil {
		log.Fatalf("could not migrate: %v", err)
	}

	m.Run()
}

type getBanksByISO2CodeTestCase struct {
	name                string
	iso2Code            string
	country             string
	expected            bool
	banksExpectedLength int
}

func runGetBanksByISO2CodeTests(t *testing.T, testCases []getBanksByISO2CodeTestCase) {
	db := GetDb()
	srv := database.New(db)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Setup()
			banks, err := srv.GetBanksByISO2Code(tc.iso2Code)
			if tc.expected {
				if err != nil {
					t.Fatalf("Name: %v, expected nil, got %v", tc.name, err)
				}
				if banks.ISO2Code != tc.iso2Code {
					t.Fatalf("Name: %v, expected %v, got %v", tc.name, tc.iso2Code, banks.ISO2Code)
				}

				if banks.Country != tc.country {
					t.Fatalf("Name: %v, expected %v, got %v", tc.name, tc.country, banks.Country)
				}

				if len(banks.Banks) != tc.banksExpectedLength {
					t.Fatalf("Name: %v, expected %v banks, got %v", tc.name, tc.banksExpectedLength, len(banks.Banks))
				}
			} else {
				if err == nil {
					t.Fatalf("Name: %v, expected error, got nil", tc.name)
				}
			}
		})
	}

}

func TestGetBanksByISO2Code(t *testing.T) {
	testCases := []getBanksByISO2CodeTestCase{
		{"Valid ISO2Code PL", "PL", "POLAND", true, 4},
		{"Valid ISO2Code US", "US", "UNITED STATES", true, 1},
		{"Valid ISO2Code NOT USED", "NT", "NOT USED COUNTRY", true, 0},
		{"Invalid ISO2Code", "XX", "INVALID", false, 0},
	}
	runGetBanksByISO2CodeTests(t, testCases)
}

type getBankBySWIFTCodeTestCase struct {
	name                   string
	swiftCode              string
	expected               bool
	branchesExpectedLength int
}

func runGetBankBySWIFTCodeTests(t *testing.T, testCases []getBankBySWIFTCodeTestCase) {
	db := GetDb()
	srv := database.New(db)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Setup()
			bank, err := srv.GetBankBySwiftCode(tc.swiftCode)
			if tc.expected {
				if err != nil {
					t.Fatalf("Name: %v, expected nil, got %v", tc.name, err)
				}
				if bank.SWIFTCode != tc.swiftCode {
					t.Fatalf("Name: %v, expected %v, got %v", tc.name, tc.swiftCode, bank.SWIFTCode)
				}
				if bank.IsHeadquarterBank() {
					if len(bank.Branches) != tc.branchesExpectedLength {
						t.Fatalf("Name: %v, expected %v branches, got %v", tc.name, tc.branchesExpectedLength, len(bank.Branches))
					}
				} else {
					if bank.Branches != nil {
						t.Fatalf("Name: %v, expected nil branches, got %v", tc.name, bank.Branches)
					}
				}

			} else {
				if err == nil {
					t.Fatalf("Name: %v, expected error, got nil", tc.name)
				}
			}
		})
	}
}

func TestGetBankBySWIFTCode(t *testing.T) {
	testCases := []getBankBySWIFTCodeTestCase{
		{"HQ bank with 2 branches", "BREXPLPWXXX", true, 2},
		{"HQ bank without branches", "AAISALTRXXX", true, 0},
		{"Branch bank", "ALBPPLP1BMW", true, 0},
		{"Not existing SWIFT code", "INVALID", false, 0},
	}
	runGetBankBySWIFTCodeTests(t, testCases)
}

type addBankFromRequestTestCase struct {
	name     string
	request  models.CreateBankRequest
	expected bool
	hqLinked bool
}

func runAddBankFromRequestTests(t *testing.T, testCases []addBankFromRequestTestCase) {
	db := GetDb()
	srv := database.New(db)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Setup()
			err := srv.AddBankFromRequest(tc.request)
			if tc.expected {
				if err != nil {
					t.Fatalf("Name: %v, expected nil, got %v", tc.name, err)
				}

				var town models.BankTown
				if err := db.Where("town = ?", tc.request.TownName).First(&town).Error; err != nil {
					t.Fatalf("Expected nil, got %v", err)
				}

				var address models.BankAddress
				if err := db.Where("address = ? AND town_id = ?", tc.request.Address, town.ID).First(&address).Error; err != nil {
					t.Fatalf("Expected nil, got %v", err)
				}

				var country models.BankCountry
				if err := db.Where("iso2_code = ? AND country_name = ?", tc.request.ISO2Code, tc.request.CountryName).First(&country).Error; err != nil {
					t.Fatalf("Expected nil, got %v", err)
				}

				var name models.BankName
				if err := db.Where("name = ?", tc.request.BankName).First(&name).Error; err != nil {
					t.Fatalf("Expected nil, got %v", err)
				}

				var codeType models.CodeType
				if err := db.Where("code_type = ?", tc.request.CodeType).First(&codeType).Error; err != nil {
					t.Fatalf("Expected nil, got %v", err)
				}

				var timeZone models.TimeZone
				if err := db.Where("time_zone = ?", tc.request.TimeZone).First(&timeZone).Error; err != nil {
					t.Fatalf("Expected nil, got %v", err)
				}

				var bank models.Bank
				bankQuery := db.Where("swift_code = ? AND code_type_id = ? AND name_id = ? AND address_id = ? AND country_id = ? AND time_zone_id = ?",
					tc.request.SWIFTCode, codeType.ID, name.ID, address.ID, country.ID, timeZone.ID)

				if tc.hqLinked {
					bankQuery = bankQuery.Where("headquarter_id IS NOT NULL")
				} else {
					bankQuery = bankQuery.Where("headquarter_id IS NULL")
				}

				if err := bankQuery.First(&bank).Error; err != nil {
					t.Fatalf("Expected nil, got %v", err)
				}

				if bank.IsHeadquarterBank() {
					var branches []models.Bank
					if err := db.Where("headquarter_id = ?", bank.ID).Find(&branches).Error; err != nil {
						t.Fatalf("Expected nil, got %v", err)
					}
					for _, branch := range branches {
						if strings.HasPrefix(branch.SWIFTCode, bank.SWIFTCode[:8]) {
							t.Fatalf("Expected nil, got %v", err)
						}
					}
				} else {
					if tc.hqLinked {
						var headquarter models.Bank
						if err := db.Where("swift_code = ? AND ID = ?", bank.SWIFTCode[:8]+"XXX", bank.HeadquarterID).First(&headquarter).Error; err != nil {
							t.Fatalf("Expected nil, got %v", err)
						}
					}
				}

			} else {
				if err == nil {
					t.Fatalf("Name: %v, expected error, got nil", tc.name)
				}
			}
		})
	}
}

func TestAddBankFromRequest(t *testing.T) {

	testCases := []addBankFromRequestTestCase{
		{
			"Correct HQ bank",
			models.CreateBankRequest{
				Address:       "Test Address",
				BankName:      "Test Bank",
				ISO2Code:      "PL",
				CountryName:   "POLAND",
				SWIFTCode:     "TESTTESTXXX",
				CodeType:      "BIC",
				TownName:      "Test Town",
				TimeZone:      "CET",
				IsHeadquarter: true,
			},
			true,
			false, // bank is HQ
		},
		{
			"Correct branch bank with existing HQ",
			models.CreateBankRequest{
				Address:       "Test Address",
				BankName:      "Test Bank",
				ISO2Code:      "PL",
				CountryName:   "POLAND",
				SWIFTCode:     "BREXPLPWNOT", // bank BREXPLPWXXX in mock_database_data.go
				CodeType:      "BIC",
				TownName:      "Test Town",
				TimeZone:      "CET",
				IsHeadquarter: false,
			},
			true,
			true,
		},
		{
			"Correct branch bank without existing HQ",
			models.CreateBankRequest{
				Address:       "Test Address",
				BankName:      "Test Bank",
				ISO2Code:      "PL",
				CountryName:   "POLAND",
				SWIFTCode:     "TESTTESTNOT",
				CodeType:      "BIC",
				TownName:      "Test Town",
				TimeZone:      "CET",
				IsHeadquarter: false,
			},
			true,
			false,
		},
		{
			"With invalid ISO2Code",
			models.CreateBankRequest{
				Address:       "Test Address",
				BankName:      "Test Bank",
				ISO2Code:      "XXW",
				CountryName:   "INVALID",
				SWIFTCode:     "TESTTESTXXX",
				CodeType:      "BIC",
				TownName:      "Test Town",
				TimeZone:      "CET",
				IsHeadquarter: true,
			},
			false,
			false,
		},
		{
			"With invalid SWIFTCode",
			models.CreateBankRequest{
				Address:       "Test Address",
				BankName:      "Test Bank",
				ISO2Code:      "PL",
				CountryName:   "POLAND",
				SWIFTCode:     "INVALID",
				CodeType:      "BIC",
				TownName:      "Test Town",
				TimeZone:      "CET",
				IsHeadquarter: true,
			},
			false,
			false,
		},
		{
			"Missing not required field",
			models.CreateBankRequest{
				Address:       "Test Address",
				BankName:      "Test Bank",
				ISO2Code:      "PL",
				CountryName:   "POLAND",
				SWIFTCode:     "TESTTESTXXX",
				IsHeadquarter: true,
			},
			true,
			false,
		},
	}
	runAddBankFromRequestTests(t, testCases)
}

type deleteWithBankCheck struct {
	name         string
	expectedType interface{}
	id           int
	expected     bool
}

type deleteBankBySwiftCodeTestCase struct {
	name       string
	swiftCode  string
	deleteWith []deleteWithBankCheck
	expected   bool
}

func runDeleteBankBySWIFTCodeTests(t *testing.T, testCases []deleteBankBySwiftCodeTestCase) {
	db := GetDb()
	srv := database.New(db)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Setup()
			err := srv.DeleteBankBySwiftCode(tc.swiftCode)
			if tc.expected {
				if err != nil {
					t.Fatalf("Name: %v, expected nil, got %v", tc.name, err)
				}

				var bank models.Bank
				if err := db.Where("swift_code = ?", tc.swiftCode).First(&bank).Error; err == nil {
					t.Fatalf("Name: %v, expected error, got nil", tc.name)
				}

				for _, check := range tc.deleteWith {
					err := db.Where("id = ?", check.id).First(check.expectedType).Error
					if check.expected && err == nil {
						t.Fatalf("Name: %v with %v, expected error, got nil", tc.name, check.name)
					}
					if !check.expected && err != nil {
						t.Fatalf("Name: %v with %v, expected nil, got %v", tc.name, check.name, err)
					}
				}

			} else {
				if err == nil {
					t.Fatalf("Name: %v, expected error, got nil", tc.name)
				}
			}
		})
	}
}

func TestDeleteBankBySwiftCode(t *testing.T) {

	testCases := []deleteBankBySwiftCodeTestCase{
		{
			"Delete HQ bank with 2 branches with foreign key used in other bank",
			"BREXPLPWXXX", // bank BREXPLPWXXX in mock_database_data.go
			[]deleteWithBankCheck{
				{"CodeType", &models.CodeType{}, 1, false},
				{"NameBank", &models.BankName{}, 1, false},
				{"Address", &models.BankTown{}, 1, false},
				{"Country", &models.BankCountry{}, 1, false},
				{"TimeZone", &models.TimeZone{}, 1, false},
			},
			true,
		},
		{
			"Delete HQ bank without branches",
			"AAISALTRXXX", // bank AAISALTRXXX in mock_database_data.go
			[]deleteWithBankCheck{
				{"CodeType", &models.CodeType{}, 2, true},
				{"NameBank", &models.BankName{}, 2, true},
				{"Address", &models.BankTown{}, 1, false},
				{"Country", &models.BankCountry{}, 2, true},
				{"TimeZone", &models.TimeZone{}, 2, true},
			},
			true,
		},
		{
			"Delete branch bank",
			"BREXPLPWWRO", // bank ALBPPLP1BMW in mock_database_data.go
			[]deleteWithBankCheck{
				{"CodeType", &models.CodeType{}, 1, false},
				{"NameBank", &models.BankName{}, 1, false},
				{"Address", &models.BankTown{}, 2, false},
				{"Country", &models.BankCountry{}, 1, false},
				{"TimeZone", &models.TimeZone{}, 1, false},
			},
			true,
		},
		{
			"Not existing SWIFT code",
			"TESTTESTNOT",
			[]deleteWithBankCheck{},
			false,
		},
	}

	runDeleteBankBySWIFTCodeTests(t, testCases)
}
