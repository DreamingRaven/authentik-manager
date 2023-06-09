package utils

import (
	"database/sql"
	"fmt"
	"strconv"

	// driver package for postgresql just needs import
	_ "github.com/lib/pq"
)

// SQLConnect gets and test a basic SQL connection to our postgres database specifically
func SQLConnect(config *SQLConfig) (*sql.DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, strconv.Itoa(config.Port), config.User, config.Password, config.DBName, config.SSLMode)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// SQLConfig the sql connection args for our postgresql db connection
type SQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}
