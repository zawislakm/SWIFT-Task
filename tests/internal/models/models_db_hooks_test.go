package models_test

import (
	"SWIFT-Remitly/internal/models"
	database_test "SWIFT-Remitly/tests/internal/database"
	"errors"
	"gorm.io/gorm"
	"strings"
	"testing"
)

type beforeDeleteTestCase struct {
	testName  string
	valueName string
	expected  error
}

func runBeforeDeleteTests(t *testing.T, testCases []beforeDeleteTestCase, findAndDeleteFunc func(*gorm.DB, string) error) {
	db := database_test.GetDb()
	for _, tc := range testCases {
		database_test.Setup()
		t.Run(tc.testName, func(t *testing.T) {
			err := findAndDeleteFunc(db, tc.valueName)
			if tc.expected == nil {
				if err != nil {
					t.Fatalf("Name: %s, expected nil, got %v", tc.testName, err)
				}
			} else {
				var targetErr *models.ErrInUse
				if !errors.As(err, &targetErr) {
					t.Fatalf("Name: %s ,expected %T, got %v", tc.testName, tc.expected, err)
				}
			}
		})
	}
}

func TestBeforeDeleteTown(t *testing.T) {
	testCases := []beforeDeleteTestCase{
		{"NotUsedTown", "NotUsedTown", nil},
		{"UsedInMultipleAddressesTown", "Town1", &models.ErrInUse{}},
		{"UsedInOneAddressTown", "Town2", &models.ErrInUse{}},
		{"TownUsedInNotUsedAddress", "TownUsedInNotUsedAddress", &models.ErrInUse{}},
	}

	findAndDeleteTown := func(db *gorm.DB, valueName string) error {
		var town models.BankTown
		db.Where("town = ?", valueName).First(&town)
		return db.Unscoped().Where("id = ?", town.ID).Delete(&town).Error
	}

	runBeforeDeleteTests(t, testCases, findAndDeleteTown)
}

func TestBeforeDeleteTimeZone(t *testing.T) {
	testCases := []beforeDeleteTestCase{
		{"NotUsedTimeZone", "NotUsedTimeZone", nil},
		{"UsedInMultipleBanksTimeZone", "Timezone1", &models.ErrInUse{}},
		{"UsedInOneBankTimeZone", "Timezone2", &models.ErrInUse{}},
	}

	findAndDeleteTimeZone := func(db *gorm.DB, valueName string) error {
		var timeZone models.TimeZone
		db.Where("time_zone = ?", valueName).First(&timeZone)
		return db.Unscoped().Where("id = ?", timeZone.ID).Delete(&timeZone).Error
	}

	runBeforeDeleteTests(t, testCases, findAndDeleteTimeZone)
}

func TestBeforeDeleteBankNames(t *testing.T) {
	testCases := []beforeDeleteTestCase{
		{"NotUsedBankName", "NotUsedBankName", nil},
		{"UsedInMultipleBanksBankName", "UsedInMultipleBanks", &models.ErrInUse{}},
		{"UsedInOneBankBankName", "BankName2", &models.ErrInUse{}},
	}

	findAndDeleteBankName := func(db *gorm.DB, valueName string) error {
		var bankName models.BankName
		db.Where("name = ?", valueName).First(&bankName)
		return db.Unscoped().Where("id = ?", bankName.ID).Delete(&bankName).Error
	}

	runBeforeDeleteTests(t, testCases, findAndDeleteBankName)
}

func TestBeforeDeleteCodeTypes(t *testing.T) {
	testCases := []beforeDeleteTestCase{
		{"NotUsedCodeType", "NotUsedCodeType", nil},
		{"UsedInMultipleBanksCodeType", "CodeType1", &models.ErrInUse{}},
		{"UsedInOneBankCodeType", "CodeType2", &models.ErrInUse{}},
	}

	findAndDeleteCodeType := func(db *gorm.DB, valueName string) error {
		var codeType models.CodeType
		db.Where("code_type = ?", valueName).First(&codeType)
		return db.Unscoped().Where("id = ?", codeType.ID).Delete(&codeType).Error
	}

	runBeforeDeleteTests(t, testCases, findAndDeleteCodeType)
}

func TestBeforeDeleteBankCountries(t *testing.T) {
	testCasesByCountry := []beforeDeleteTestCase{
		{"NotUsedCountry-Country", "NOT USED COUNTRY", nil},
		{"UsedInMultipleBanksCountry-Country", "POLAND", &models.ErrInUse{}},
		{"UsedInOneBankCountry-Country", "UNITED STATES", &models.ErrInUse{}},
	}

	findAndDeleteBankCountryByCountry := func(db *gorm.DB, valueName string) error {
		var bankCountry models.BankCountry
		db.Where("country_name = ?", valueName).First(&bankCountry)
		return db.Unscoped().Where("id = ?", bankCountry.ID).Delete(&bankCountry).Error
	}
	runBeforeDeleteTests(t, testCasesByCountry, findAndDeleteBankCountryByCountry)

	testCasesByISO2Code := []beforeDeleteTestCase{
		{"NotUsedCountry-ISO2", "NT", nil},
		{"UsedInMultipleBanksCountry-ISO2", "PL", &models.ErrInUse{}},
		{"UsedInOneBankCountry-IS02", "US", &models.ErrInUse{}},
	}

	findAndDeleteBankCountryByISO2Code := func(db *gorm.DB, valueName string) error {
		var bankCountry models.BankCountry
		db.Where("iso2_code = ?", valueName).First(&bankCountry)
		return db.Unscoped().Where("id = ?", bankCountry.ID).Delete(&bankCountry).Error
	}
	runBeforeDeleteTests(t, testCasesByISO2Code, findAndDeleteBankCountryByISO2Code)

}

func TestBeforeDeleteAddress(t *testing.T) {
	testCases := []beforeDeleteTestCase{
		{"NotUsedAddressWithNotUsedTown", "NotUsedAddressWithNotUsedTown", nil},
		{"NotUsedAddressWithUsedTown", "NotUsedAddressWithUsedTown", nil},
		{"UsedAddress with Town used multiple times", "Address1", &models.ErrInUse{}},
		{"UsedAddress with Town used one time", "Address2", &models.ErrInUse{}},
	}

	findAndDeleteAddress := func(db *gorm.DB, valueName string) error {
		var address models.BankAddress
		db.Where("address = ?", valueName).First(&address)
		return db.Unscoped().Where("id = ?", address.ID).Delete(&address).Error
	}

	runBeforeDeleteTests(t, testCases, findAndDeleteAddress)
}

type beforeCreateTestCase struct {
	testName    string
	input       interface{} // Może to być models.BankCountry lub models.BankName
	expectError bool
	expectedErr error
}

func runBeforeCreateTests(t *testing.T, testCases []beforeCreateTestCase) {
	db := database_test.GetDb()
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			database_test.Setup()
			err := db.Create(tc.input).Error

			if tc.expectError {
				if err == nil {
					t.Fatalf("Name: %s, expected error, got nil", tc.testName)
				}
				var invalidData *models.ErrInvalidData
				if !errors.As(err, &invalidData) {
					t.Fatalf("Name: %s, expected %T, got %v", tc.testName, invalidData, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Name: %s, expected nil, got %v", tc.testName, err)
				}
			}
		})
	}
}

func TestBeforeCreateBankCountry(t *testing.T) {
	testCases := []beforeCreateTestCase{
		{
			testName:    "ValidBankCountry",
			input:       &models.BankCountry{ISO2Code: "OK", CountryName: "OK COUNTRY"},
			expectError: false,
		},
		{
			testName:    "InvalidISO2Length",
			input:       &models.BankCountry{ISO2Code: "OKK", CountryName: "TOOLONG COUNTRY"},
			expectError: true,
			expectedErr: &models.ErrInvalidData{},
		},
		{
			testName:    "InvalidISO2Case",
			input:       &models.BankCountry{ISO2Code: "ok", CountryName: "TOOLONG COUNTRY"},
			expectError: true,
			expectedErr: &models.ErrInvalidData{},
		},
		{
			testName:    "InvalidCountryLength",
			input:       &models.BankCountry{ISO2Code: "OK", CountryName: ""},
			expectError: true,
			expectedErr: &models.ErrInvalidData{},
		},
		{
			testName:    "InvalidCountryCase",
			input:       &models.BankCountry{ISO2Code: "OK", CountryName: "wrong country"},
			expectError: true,
			expectedErr: &models.ErrInvalidData{},
		},
	}

	runBeforeCreateTests(t, testCases)
}

func TestBeforeCreateBankName(t *testing.T) {
	testCases := []beforeCreateTestCase{
		{
			testName:    "ValidBankName",
			input:       &models.BankName{Name: "OK NAME"},
			expectError: false,
		},
		{
			testName:    "InvalidBankName",
			input:       &models.BankName{Name: ""},
			expectError: true,
			expectedErr: &models.ErrInvalidData{},
		},
	}

	runBeforeCreateTests(t, testCases)
}

func TestBeforeCreateBank(t *testing.T) {
	testCases := []beforeCreateTestCase{
		{
			testName:    "ValidBank",
			input:       &models.Bank{SWIFTCode: "TESTTESTXXX", CodeTypeID: 1, NameID: 1, AddressID: 1, CountryID: 1, TimeZoneID: 1},
			expectError: false,
		},
		{
			testName:    "InvalidSWIFTCodeLength",
			input:       &models.Bank{SWIFTCode: "BREXPLP", CodeTypeID: 1, NameID: 1, AddressID: 1, CountryID: 1, TimeZoneID: 1},
			expectError: true,
			expectedErr: &models.ErrInvalidData{},
		},
		{
			testName:    "InvalidSWIFTCodeCase",
			input:       &models.Bank{SWIFTCode: "brexplpwxxx", CodeTypeID: 1, NameID: 1, AddressID: 1, CountryID: 1, TimeZoneID: 1},
			expectError: true,
			expectedErr: &models.ErrInvalidData{},
		},
	}
	runBeforeCreateTests(t, testCases)
}

func TestAfterCreateBankHQBank(t *testing.T) {
	db := database_test.GetDb()
	database_test.Setup()

	HQBank := models.Bank{
		SWIFTCode:  "ALBPPLP1XXX", // ALBPPLP1BMW in mock data, after creation of HQBank ALBPPLP1BMW will have linked HeadquarterID
		CodeTypeID: 1,
		NameID:     1,
		AddressID:  1,
		CountryID:  1,
		TimeZoneID: 1,
	}

	if err := db.Create(&HQBank).Error; err != nil {
		t.Fatalf("Error during creating HQBank: %v", err)
	}

	var branchBanks []models.Bank
	if err := db.Where("headquarter_id = ?", HQBank.ID).Find(&branchBanks).Error; err != nil {
		t.Fatalf("Error during fetching branch banks: %v", err)
	}

	if len(branchBanks) != 1 {
		t.Fatalf("Expected 1 branch banks, got %v", len(branchBanks))
	}

	for _, branchBank := range branchBanks {
		if !strings.HasPrefix(branchBank.SWIFTCode, HQBank.SWIFTCode[:8]) {
			t.Fatalf("Expected branch bank SWIFT code to start with %v, got %v", HQBank.SWIFTCode[:8], branchBank.SWIFTCode)
		}
	}

}

func TestAfterCreateBankBranchBank(t *testing.T) {
	db := database_test.GetDb()
	database_test.Setup()

	branchBank := models.Bank{
		SWIFTCode:  "BREXPLPWTTT", // BREXPLPWXXX in mock data, after creation of branchBank will get HQ linked
		CodeTypeID: 1,
		NameID:     1,
		AddressID:  1,
		CountryID:  1,
		TimeZoneID: 1,
	}

	if err := db.Create(&branchBank).Error; err != nil {
		t.Fatalf("Error during creating branchBank: %v", err)
	}

	if branchBank.HeadquarterID == nil {
		t.Fatalf("Expected branch bank to have linked HQ, got %v", branchBank.HeadquarterID)
	}

	var HQBank models.Bank
	if err := db.Where("id = ?", *branchBank.HeadquarterID).First(&HQBank).Error; err != nil {
		t.Fatalf("Error during fetching HQBank: %v", err)
	}

	if !strings.HasPrefix(branchBank.SWIFTCode, HQBank.SWIFTCode[:8]) {
		t.Fatalf("Expected branch bank SWIFT code to start with %v, got %v", HQBank.SWIFTCode[:8], branchBank.SWIFTCode)
	}

	var branchBanks []models.Bank
	if err := db.Where("headquarter_id = ?", *branchBank.HeadquarterID).Find(&branchBanks).Error; err != nil {
		t.Fatalf("Error during fetching branch banks: %v", err)
	}

	if len(branchBanks) != 3 {
		t.Fatalf("Expected 3 branch banks, got %v", len(branchBanks))
	}

	for _, branchBank := range branchBanks {
		if !strings.HasPrefix(branchBank.SWIFTCode, HQBank.SWIFTCode[:8]) {
			t.Fatalf("Expected branch bank SWIFT code to start with %v, got %v", HQBank.SWIFTCode[:8], branchBank.SWIFTCode)
		}
	}

}
