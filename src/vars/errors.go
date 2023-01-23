package vars

import "errors"

// ErrNoIPv4AddressAssigned may be thrown if a function needs an IPv4 address on at least one network interface,
// but no IPv4 addresses are assigned
var ErrNoIPv4AddressAssigned = errors.New("no accessible interface has a IPv4 address assigned to it")

// ErrEnvironmentVariableNotFound will be thrown if an environment shall be read by the utility function but the
// variable is not populated
var ErrEnvironmentVariableNotFound = errors.New("the specified environment variable was not populated")

var ErrHttpErrorNotFound = errors.New("the supplied error code does not match any configured http errors")
