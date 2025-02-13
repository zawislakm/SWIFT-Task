package models

type TimeZone struct {
	ID       uint   `gorm:"primaryKey"`
	TimeZone string `gorm:"unique;not null"`
}

type BankCountry struct {
	ID          uint   `gorm:"primaryKey"`
	ISO2Code    string `gorm:"index:idx_country_iso2_code,unique;not null"`
	CountryName string `gorm:"index:idx_country_iso2_code,unique;not null"`
}

type BankName struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

type CodeType struct {
	ID       uint   `gorm:"primaryKey"`
	CodeType string `gorm:"unique;not null"`
}

type BankTown struct {
	ID   uint   `gorm:"primaryKey"`
	Town string `gorm:"unique;not null"`
}

type BankAddress struct {
	ID      uint     `gorm:"primaryKey"`
	Address string   `gorm:"index:idx_address_town,unique;not null"`
	TownID  uint     `gorm:"index:idx_address_town,unique;not null"`
	Town    BankTown `gorm:"foreignKey:TownID"`
}

type Bank struct {
	ID            uint   `gorm:"primaryKey"`
	SWIFTCode     string `gorm:"unique;not null"`
	CodeTypeID    uint
	CodeType      CodeType `gorm:"foreignKey:CodeTypeID"`
	NameID        uint
	Name          BankName `gorm:"foreignKey:NameID"`
	AddressID     uint
	Address       BankAddress `gorm:"foreignKey:AddressID"`
	CountryID     uint
	Country       BankCountry `gorm:"foreignKey:CountryID"`
	TimeZoneID    uint
	TimeZone      TimeZone `gorm:"foreignKey:TimeZoneID"`
	HeadquarterID *uint
	Headquarter   *Bank  `gorm:"foreignKey:HeadquarterID"`
	Branches      []Bank `gorm:"-"`
}

type BankDataCSV struct {
	Address     string `csv:"ADDRESS"`
	Name        string `csv:"NAME"`
	ISO2Code    string `csv:"COUNTRY ISO2 CODE"`
	CountryName string `csv:"COUNTRY NAME"`
	SwiftCode   string `csv:"SWIFT CODE"`
	CodeType    string `csv:"CODE TYPE"`
	TownName    string `csv:"TOWN NAME"`
	TimeZone    string `csv:"TIME ZONE"`
}

type CountrySWIFTCode struct {
	ISO2Code string `json:"iso2Code"`
	Country  string `json:"country"`
	Banks    []Bank `json:"swiftCodes"`
}

type CreateBankRequest struct {
	Address       string `json:"address" csv:"ADDRESS"`
	BankName      string `json:"bankName" csv:"NAME"`
	ISO2Code      string `json:"countryISO2" csv:"COUNTRY ISO2 CODE"`
	CountryName   string `json:"countryName" csv:"COUNTRY NAME"`
	SWIFTCode     string `json:"swiftCode" csv:"SWIFT CODE"`
	CodeType      string `json:"-" csv:"CODE TYPE"`
	TownName      string `json:"-" csv:"TOWN NAME"`
	TimeZone      string `json:"-" csv:"TIME ZONE"`
	IsHeadquarter bool   `json:"isHeadquarter" csv:"-"`
}

type Response struct {
	Success bool     `json:"success"`
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}
