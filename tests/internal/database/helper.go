package database

import (
	"SWIFT-Remitly/internal/models"
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type service struct {
	Db        *gorm.DB
	Container testcontainers.Container
}

var (
	dbInstance *service
)

// StartPostgresContainer sets up a PostgreSQL container and returns a database connection and a function to close the container.
func StartPostgresContainer(ctx context.Context) (*gorm.DB, testcontainers.Container, error) {
	containerRequest := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerRequest,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get container host: %w", err)
	}
	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get container port: %w", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=testdb sslmode=disable", host, port.Port())

	testLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Silent,
		})

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: testLogger})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("PostgreSQL container started successfully")

	dbInstance = &service{
		Db:        db,
		Container: postgresContainer,
	}

	return dbInstance.Db, dbInstance.Container, nil
}

func GetDb() *gorm.DB {
	if dbInstance == nil {
		db, _, err := StartPostgresContainer(context.Background())
		if err != nil {
			log.Fatalf("could not start postgres container: %v", err)
		}
		return db
	}
	return dbInstance.Db
}

func Setup() {
	db := GetDb()
	if err := MockData(db); err != nil {
		log.Fatalf("could not insert test data: %v", err)
	}
}

func migrate(db *gorm.DB) error {
	err := db.Migrator().DropTable(&models.TimeZone{}, &models.BankCountry{}, &models.BankName{}, &models.CodeType{}, &models.BankTown{}, &models.BankAddress{}, &models.Bank{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&models.TimeZone{}, &models.BankCountry{}, &models.BankName{}, &models.CodeType{}, &models.BankTown{}, &models.BankAddress{}, &models.Bank{})
	if err != nil {
		return err
	}
	return nil
}
