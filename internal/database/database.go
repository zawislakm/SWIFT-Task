package database

import (
	"SWIFT-Remitly/internal/models"
	"context"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

// Service represents a service that interacts with a database.
type Service interface {

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// GetBankBySwiftCode retrieves the bank data from the database based on the SWIFT code.
	// It returns the bank data and an error if the bank data cannot be retrieved.
	GetBankBySwiftCode(swiftCode string) (models.Bank, error)

	// GetBanksByISO2Code retrieves the bank data from the database based on the ISO2 code.
	// It returns the bank data and an error if the bank data cannot be retrieved.
	GetBanksByISO2Code(iso2Code string) (models.CountrySWIFTCode, error)

	// AddBankFromRequest adds the bank data to the database.
	// It returns an error if the bank data cannot be added.
	AddBankFromRequest(requestData models.CreateBankRequest) error

	// DeleteBankBySwiftCode the bank data from the database based on the SWIFT code.
	// It returns an error if the bank data cannot be removed.
	DeleteBankBySwiftCode(swiftCode string) error
}

type service struct {
	db *gorm.DB
}

var (
	database   = os.Getenv("POSTGRES_DB")
	password   = os.Getenv("POSTGRES_PASSWORD")
	username   = os.Getenv("POSTGRES_USER")
	port       = os.Getenv("POSTGRES_DB_PORT")
	host       = os.Getenv("POSTGRES_DB_HOST")
	schema     = os.Getenv("POSTGRES_DB_SCHEMA")
	dbInstance *service
)

func New(dbIn *gorm.DB) Service {
	if dbIn != nil {
		return &service{
			db: dbIn,
		}
	}
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable search_path=%s", host, username, password, database, port, schema)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer)
		logger.Config{
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		})

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         newLogger,
		TranslateError: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}

	if err = dbInstance.migrate(); err != nil {
		log.Fatal(err)
	}

	return dbInstance
}

func (s *service) migrate() error {
	s.db.Logger.Info(context.Background(), "Migrating the database")

	s.db.Logger.Info(context.Background(), "Dropping tables")
	err := s.db.Migrator().DropTable(&models.TimeZone{}, &models.BankCountry{}, &models.BankName{}, &models.CodeType{}, &models.BankTown{}, &models.BankAddress{}, &models.Bank{})
	if err != nil {
		s.db.Logger.Error(context.Background(), "Error during dropping tables: "+err.Error())
		return err
	}

	s.db.Logger.Info(context.Background(), "Auto migrating tables")
	err = s.db.AutoMigrate(&models.TimeZone{}, &models.BankCountry{}, &models.BankName{}, &models.CodeType{}, &models.BankTown{}, &models.BankAddress{}, &models.Bank{})
	if err != nil {
		s.db.Logger.Error(context.Background(), "Error during auto migrating tables: "+err.Error())
		return err
	}
	return nil
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	s.db.Logger.Info(context.Background(), "Disconnecting from database: "+database)
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	log.Printf("Disconnected from database: %s", database)
	return sqlDB.Close()
}

// GetBanksByISO2Code retrieves the banks data from the database based on the ISO2 code.
func (s *service) GetBanksByISO2Code(iso2Code string) (models.CountrySWIFTCode, error) {
	s.db.Logger.Info(context.Background(), "Retrieving banks data from the database by ISO2 code")

	bankCountry, err := s.getCountryByISO2Code(iso2Code)
	if err != nil {
		s.db.Logger.Error(context.Background(), "Error during retrieving country by ISO2 code: "+err.Error())
		return models.CountrySWIFTCode{}, err
	}

	var banks []models.Bank

	if err := s.db.
		Preload("Name").
		Preload("Address").
		Preload("Country", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, iso2_code")
		}).
		Where("country_id = ?", bankCountry.ID).
		Find(&banks).Error; err != nil {
		s.db.Logger.Error(context.Background(), "Error during retrieving banks by ISO2 code: "+err.Error())
		return models.CountrySWIFTCode{}, err
	}

	var country models.CountrySWIFTCode
	country.ISO2Code = bankCountry.ISO2Code
	country.Country = bankCountry.CountryName
	country.Banks = banks

	return country, nil
}

// GetBankBySwiftCode retrieves the bank data from the database based on the SWIFT code.
func (s *service) GetBankBySwiftCode(swiftCode string) (models.Bank, error) {
	s.db.Logger.Info(context.Background(), "Retrieving bank data from the database by SWIFT code")

	var bank models.Bank
	if err := s.db.
		Preload("Name").
		Preload("Address").
		Preload("Country").
		Where("swift_code = ?", swiftCode).
		First(&bank).Error; err != nil {
		s.db.Logger.Error(context.Background(), "Error during retrieving bank by SWIFT code: "+err.Error())
		return models.Bank{}, err
	}

	if bank.IsHeadquarterBank() {
		branches, err := s.getHeadquarterBranches(bank.ID)
		if err != nil {
			s.db.Logger.Error(context.Background(), "Error during retrieving branches of a headquarters bank: "+err.Error())
			return models.Bank{}, err
		}
		bank.Branches = branches
	}
	return bank, nil
}

// AddBankFromRequest adds the bank data to the database.
func (s *service) AddBankFromRequest(requestData models.CreateBankRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		tx.Logger.Info(tx.Statement.Context, "Adding bank data to the database")

		var timeZone models.TimeZone
		if err := tx.
			Where("time_zone = ?", requestData.TimeZone).
			FirstOrCreate(&timeZone, models.TimeZone{TimeZone: requestData.TimeZone}).Error; err != nil {
			tx.Logger.Error(tx.Statement.Context, "Error during adding bank: "+err.Error())
			return err
		}

		var country models.BankCountry
		if err := tx.
			Where("iso2_code = ? AND country_name = ?", requestData.ISO2Code, requestData.CountryName).
			FirstOrCreate(&country, models.BankCountry{ISO2Code: requestData.ISO2Code, CountryName: requestData.CountryName}).Error; err != nil {
			tx.Logger.Error(tx.Statement.Context, "Error during adding bank: "+err.Error())
			return err
		}

		var name models.BankName
		if err := tx.
			Where("name = ?", requestData.BankName).
			FirstOrCreate(&name, models.BankName{Name: requestData.BankName}).Error; err != nil {
			tx.Logger.Error(tx.Statement.Context, "Error during adding bank: "+err.Error())
			return err
		}

		var codeType models.CodeType
		if err := tx.
			Where("code_type = ?", requestData.CodeType).
			FirstOrCreate(&codeType, models.CodeType{CodeType: requestData.CodeType}).Error; err != nil {
			tx.Logger.Error(tx.Statement.Context, "Error during adding bank: "+err.Error())
			return err
		}

		var town models.BankTown
		if err := tx.
			Where("town = ?", requestData.TownName).
			FirstOrCreate(&town, models.BankTown{Town: requestData.TownName}).Error; err != nil {
			tx.Logger.Error(tx.Statement.Context, "Error during adding bank: "+err.Error())
			return err
		}

		var address models.BankAddress
		if err := tx.
			Where("address = ? AND town_id = ?", requestData.Address, town.ID).
			FirstOrCreate(&address, models.BankAddress{Address: requestData.Address, TownID: town.ID}).Error; err != nil {
			tx.Logger.Error(tx.Statement.Context, "Error during adding bank: "+err.Error())
			return err
		}

		bank := models.Bank{
			SWIFTCode:  requestData.SWIFTCode,
			CodeTypeID: codeType.ID,
			NameID:     name.ID,
			AddressID:  address.ID,
			CountryID:  country.ID,
			TimeZoneID: timeZone.ID,
		}

		if err := tx.
			Create(&bank).Error; err != nil {
			tx.Logger.Error(tx.Statement.Context, "Error during adding bank: "+err.Error())
			return err
		}

		return nil
	})
}

// DeleteBankBySwiftCode deletes the bank data from the database based on the SWIFT code.
func (s *service) DeleteBankBySwiftCode(swiftCode string) error {
	/*
		I don't like the way this is done, but I'm not sure how to fix it
		https://gorm.io/docs/associations.html#Delete-Associations
		tried this ^, but it didn't work, or I did something wrong
		specifying CASCADE in models didn't work either
	*/
	return s.db.Transaction(func(tx *gorm.DB) error {
		tx.Logger.Info(tx.Statement.Context, "Deleting bank data from the database")
		var bank models.Bank
		if err := tx.
			Preload("Address").
			Preload("CodeType").
			Preload("Country").
			Preload("Name").
			Preload("TimeZone").
			Where("swift_code = ?", swiftCode).
			First(&bank).Error; err != nil {
			tx.Logger.Error(tx.Statement.Context, "Error during deleting bank: "+err.Error())
			return err
		}

		if err := tx.Unscoped().Delete(&bank).Error; err != nil {
			tx.Logger.Error(tx.Statement.Context, "Error during deleting bank: "+err.Error())
			return err
		}

		entities := []interface{}{&bank.Address, &bank.CodeType, &bank.Country, &bank.Name, &bank.TimeZone}
		for _, entity := range entities {
			if err := tx.Unscoped().Delete(entity).Error; err != nil {
				if err := handleDeleteError(tx, err); err != nil {
					tx.Logger.Error(
						tx.Statement.Context,
						fmt.Sprintf("Error during deleting bank, error with %s entity: %s", tx.Statement.Table, err.Error()))
					return err
				}
			}
		}

		return nil
	})
}
