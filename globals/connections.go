package globals

import "database/sql"

// This file contains all globally shared connections (e.g., Databases)

// Db contains the globally available connection to the database
var Db *sql.DB
