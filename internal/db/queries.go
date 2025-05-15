package db

import (
	"io/fs"
	"log/slog"

	"github.com/qustavo/dotsql"

	"microservice/resources"
)

// This file contains the globally shared Queries variable

// Queries contains the prepared sql queries from the resources folder.
var Queries *dotsql.DotSql

func LoadQueries() error {
	slog.Debug("loading embedded database queries")
	queryFiles, err := fs.ReadDir(resources.QueryFiles, ".")
	if err != nil {
		return err
	}

	dotSqls := make([]*dotsql.DotSql, len(queryFiles))
	for idx, queryFile := range queryFiles {
		instance, err := dotsql.LoadFromFile(queryFile.Name())
		if err != nil {
			return err
		}
		dotSqls[idx] = instance
	}

	Queries = dotsql.Merge(dotSqls...)
	return nil
}
