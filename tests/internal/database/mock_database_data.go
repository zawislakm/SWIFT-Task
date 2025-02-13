package database

import (
	"SWIFT-Remitly/internal/models"
	"gorm.io/gorm"
)

var (
	sampleTowns = []models.BankTown{
		{Town: "Town1"},                    // used in 4 addresses, one of them is not used in any bank
		{Town: "Town2"},                    // used in 1 address
		{Town: "Town3"},                    // used in 1 address
		{Town: "TownUsedInNotUsedAddress"}, // used in 1 address
		{Town: "NotUsedTown"},              // not used
	}

	sampleTimeZones = []models.TimeZone{
		{TimeZone: "Timezone1"},       // used in 4 banks
		{TimeZone: "Timezone2"},       // used in 1 bank
		{TimeZone: "NotUsedTimezone"}, // not used
	}

	sampleBankNames = []models.BankName{
		{Name: "UsedInMultipleBanks"}, // used in 3 banks
		{Name: "BankName2"},           // used in 1 bank
		{Name: "BankName3"},           // used in 1 bank
		{Name: "NotUsedBankName"},     // not used
	}

	sampleCodeTypes = []models.CodeType{
		{CodeType: "CodeType1"},       // used in 4 banks
		{CodeType: "CodeType2"},       // used in 1 bank
		{CodeType: "NotUsedCodeType"}, // not used
	}

	sampleBankCountries = []models.BankCountry{
		{ISO2Code: "PL", CountryName: "POLAND"},           // used in 4 banks
		{ISO2Code: "US", CountryName: "UNITED STATES"},    // used in 1 bank
		{ISO2Code: "NT", CountryName: "NOT USED COUNTRY"}, // not used

	}

	//filled with correct data in prepareAddresses and prepareBanks
	sampleAddresses = []models.BankAddress{}
	sampleBanks     = []models.Bank{}
)

// TODO rethink this loops

func prepareTowns(db *gorm.DB) error {
	for i := range sampleTowns {
		if err := db.Raw("INSERT INTO bank_towns (town) VALUES (?) RETURNING id", sampleTowns[i].Town).
			Scan(&sampleTowns[i].ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func prepareTimeZones(db *gorm.DB) error {
	for i := range sampleTimeZones {
		if err := db.Raw("INSERT INTO time_zones (time_zone) VALUES (?) RETURNING id", sampleTimeZones[i].TimeZone).
			Scan(&sampleTimeZones[i].ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func prepareBankNames(db *gorm.DB) error {
	for i := range sampleBankNames {
		if err := db.Raw("INSERT INTO bank_names (name) VALUES (?) RETURNING id", sampleBankNames[i].Name).
			Scan(&sampleBankNames[i].ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func prepareCodeTypes(db *gorm.DB) error {
	for i := range sampleCodeTypes {
		if err := db.Raw("INSERT INTO code_types (code_type) VALUES (?) RETURNING id", sampleCodeTypes[i].CodeType).
			Scan(&sampleCodeTypes[i].ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func prepareBankCountries(db *gorm.DB) error {
	for i := range sampleBankCountries {
		if err := db.Raw("INSERT INTO bank_countries (iso2_code, country_name) VALUES (?, ?) RETURNING id", sampleBankCountries[i].ISO2Code, sampleBankCountries[i].CountryName).
			Scan(&sampleBankCountries[i].ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func prepareAddresses(db *gorm.DB) error {

	sampleAddresses = []models.BankAddress{
		{Address: "Address1", TownID: sampleTowns[0].ID},                      // used 2 times
		{Address: "Address2", TownID: sampleTowns[0].ID},                      // used 1 time
		{Address: "Address3", TownID: sampleTowns[0].ID},                      // used 0 times with town used in other addresses
		{Address: "Address3", TownID: sampleTowns[1].ID},                      // used 1 time
		{Address: "Address4", TownID: sampleTowns[2].ID},                      // used 1 time
		{Address: "NotUsedAddressWithNotUsedTown", TownID: sampleTowns[3].ID}, // not used with town not used in other addresses
		{Address: "NotUsedAddressWithUsedTown", TownID: sampleTowns[0].ID},    // not used with town used in other addresses
	}

	for i := range sampleAddresses {
		if err := db.Raw("INSERT INTO bank_addresses (address, town_id) VALUES (?, ?) RETURNING id", sampleAddresses[i].Address, sampleAddresses[i].TownID).
			Scan(&sampleAddresses[i].ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func prepareBanks(db *gorm.DB) error {
	sampleBanks = []models.Bank{
		{SWIFTCode: "BREXPLPWXXX", CodeTypeID: sampleCodeTypes[0].ID, NameID: sampleBankNames[0].ID, AddressID: sampleAddresses[0].ID, CountryID: sampleBankCountries[0].ID, TimeZoneID: sampleTimeZones[0].ID},
		{SWIFTCode: "AAISALTRXXX", CodeTypeID: sampleCodeTypes[1].ID, NameID: sampleBankNames[1].ID, AddressID: sampleAddresses[0].ID, CountryID: sampleBankCountries[1].ID, TimeZoneID: sampleTimeZones[1].ID},
		{SWIFTCode: "ALBPPLP1BMW", CodeTypeID: sampleCodeTypes[0].ID, NameID: sampleBankNames[2].ID, AddressID: sampleAddresses[4].ID, CountryID: sampleBankCountries[0].ID, TimeZoneID: sampleTimeZones[0].ID},
	}

	for i := range sampleBanks {
		if err := db.Raw("INSERT INTO banks (swift_code, code_type_id, name_id, address_id, country_id, time_zone_id) VALUES (?, ?, ?, ?, ?, ?) RETURNING id",
			sampleBanks[i].SWIFTCode, sampleBanks[i].CodeTypeID, sampleBanks[i].NameID, sampleBanks[i].AddressID, sampleBanks[i].CountryID, sampleBanks[i].TimeZoneID).
			Scan(&sampleBanks[i].ID).Error; err != nil {
			return err
		}
	}
	sampleBranches := []models.Bank{
		{SWIFTCode: "BREXPLPWWRO", CodeTypeID: sampleCodeTypes[0].ID, NameID: sampleBankNames[0].ID, AddressID: sampleAddresses[1].ID, CountryID: sampleBankCountries[0].ID, TimeZoneID: sampleTimeZones[0].ID, HeadquarterID: &sampleBanks[0].ID},
		{SWIFTCode: "BREXPLPWWAL", CodeTypeID: sampleCodeTypes[0].ID, NameID: sampleBankNames[0].ID, AddressID: sampleAddresses[2].ID, CountryID: sampleBankCountries[0].ID, TimeZoneID: sampleTimeZones[0].ID, HeadquarterID: &sampleBanks[0].ID}}

	for i := range sampleBranches {
		if err := db.Raw("INSERT INTO banks (swift_code, code_type_id, name_id, address_id, country_id, time_zone_id, headquarter_id) VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id",
			sampleBranches[i].SWIFTCode, sampleBranches[i].CodeTypeID, sampleBranches[i].NameID, sampleBranches[i].AddressID, sampleBranches[i].CountryID, sampleBranches[i].TimeZoneID, sampleBranches[i].HeadquarterID).
			Scan(&sampleBranches[i].ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func MockData(db *gorm.DB) error {
	if err := migrate(db); err != nil {
		return err
	}

	functions := []func(*gorm.DB) error{
		prepareBankNames,
		prepareTimeZones,
		prepareCodeTypes,
		prepareBankCountries,
		prepareTowns,
		prepareAddresses,
		prepareBanks,
	}

	for _, fn := range functions {
		if err := fn(db); err != nil {
			return err
		}
	}

	return nil
}
