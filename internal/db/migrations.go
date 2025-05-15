//go:build !no_migrations

package db

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"microservice/internal"
	"microservice/resources"
)

const databaseDialect = "postgres"

func MigrateDatabase() error {

	goose.SetTableName(generateVersionTableName(internal.ServiceName))
	goose.SetBaseFS(resources.DatabaseMigrations)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect(databaseDialect); err != nil {
		return err
	}

	db := stdlib.OpenDBFromPool(Pool())

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}
	return nil

}

func generateVersionTableName(serviceName string) string {
	const format = `%s_table_version`
	return fmt.Sprintf(format, strcase.ToSnake(serviceName))
}
