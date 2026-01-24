package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/stuneak/sopeko/pkg/logger"
)

var dblog = logger.NewLogger("DB")

func NewDB(driver, source string) (*sql.DB, error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		dblog("failed to open connection: %v", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		dblog("ping failed: %v", err)
		return nil, err
	}

	dblog("connected successfully")
	return db, nil
}
