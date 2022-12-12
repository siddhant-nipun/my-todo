package database

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	// source/file import is required for migration files to read
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	//"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var Todo *sqlx.DB

type SSLMode string

const (
	SSLModeEnable  SSLMode = "enable"
	SSLModeDisable SSLMode = "disable"
)

//ConnectAndMigrate function connects to database and returns error if any
func ConnectAndMigrate(host, port, databaseName, user, password string, sslMode SSLMode) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, databaseName, sslMode)
	DB, err := sqlx.Open("postgres", connStr)

	if err != nil {
		return err
	}

	err = DB.Ping()
	if err != nil {
		return err
	}
	Todo = DB
	return migrateUp(DB)
}

//migrateUp Database and handle migration logic
func migrateUp(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://database/migrations", "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// Shutdown database
func Shutdown() {
	Todo.Close()
}

// Tx provides the transaction wrapper
func Tx(fn func(tx *sqlx.Tx) error) error {
	tx, err := Todo.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start a traction: %+v", err)
	}
	defer func() {
		if err != nil {
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				logrus.Errorf("failed to rollback tx: %s", rollBackErr)
			}
		}
		if commitErr := tx.Commit(); commitErr != nil {
			logrus.Errorf("failed to commit tx: %s", commitErr)
		}
	}()
	err = fn(tx)
	return err
}

////SetupBindVars prepares the SQL statement for batch insert
//func SetupBindVars(stmt, bindVars string, length int) string {
//	bindVars += ","
//	stmt = fmt.Sprintf(stmt, strings.Repeat(bindVars, length))
//	return replaceSQL(strings.TrimSuffix(stmt, ","), "?")
//}
//
//// replaceSQL replaces the instance occurrence of any string pattern with an increasing $n based sequence
//func replaceSQL(old, searchPattern string) string {
//	tempCount := strings.Count(old, searchPattern)
//	for m := 1; m <= tempCount; m++ {
//		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
//	}
//	return old
//}
