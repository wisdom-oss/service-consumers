package connections

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
)

// DbConnection stores the pointer for the database connection this service uses
var DbConnection *sql.DB

// RedisClient stores the pointer for the client accessing the redis database
// which stores the issued ETags of responses and the respective responses
var RedisClient *redis.Client
