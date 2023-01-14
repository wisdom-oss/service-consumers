package utils

import (
	"os"

	log "github.com/sirupsen/logrus"

	"microservice/vars"
)

var logger = log.WithFields(
	log.Fields{
		"package": "utils",
	},
)

// ArrayContains takes an array of comparable values and another comparable value of the same type and checks if the
// value is present in the array
func ArrayContains[V comparable](array []V, value V) bool {
	for _, item := range array {
		if item == value {
			return true
		}
	}
	return false
}

// MapContainsKey takes a generic map and iterates through the key and searches for the lookup value.
// The lookup value needs to be of the same type as the key of the mapping
func MapContainsKey[K comparable, V any](mapping map[K]V, lookupValue K) bool {
	for key, _ := range mapping {
		if key == lookupValue {
			return true
		}
	}
	return false
}

// ReadEnvironmentVariable takes the name of an environment variable and checks its existence and returns the value
// if the variable is populated. If the variable is not populated an error will be returned
func ReadEnvironmentVariable(key string) (string, error) {
	variableValue, variableIsSet := os.LookupEnv(key)
	if variableIsSet {
		return variableValue, nil
	} else {
		return "", vars.ErrEnvironmentVariableNotFound
	}
}
