
package db

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/pressly/goose/v3"
)

func OpenAndMigrate(dsn string) (*sql.DB, error) {
    sqlDB, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, fmt.Errorf("sql.Open: %w", err)
    }
    sqlDB.SetMaxOpenConns(10)
    sqlDB.SetMaxIdleConns(5)

    if err := sqlDB.Ping(); err != nil {
        return nil, fmt.Errorf("db.Ping: %w", err)
    }

    if err := goose.SetDialect("postgres"); err != nil {
        return nil, fmt.Errorf("goose dialect: %w", err)
    }
    if err := goose.Up(sqlDB, "migrations"); err != nil {
        return nil, fmt.Errorf("goose.Up: %w", err)
    }
    log.Println("Migrations applied successfully.")
    return sqlDB, nil
}
