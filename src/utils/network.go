package utils

import (
	"fmt"
	"net"
	"time"

	"microservice/vars"
)

// PingHost tries to ping to supplied host on the supplied port by dialing to it via tcp.
// If the dial was successful in the specified timeout the function will return true.
//
// Parameters:
//  - host: the host which shall be pinged
//  - port: the port which shall be pinged on the host
//  - timeout: the timeout which is set for the ping in seconds
func PingHost(host string, port int, timeout int) bool {
	connectionTimeout := time.Duration(timeout) * time.Second
	connectionTarget := fmt.Sprintf("%s:%d", host, port)
	connection, connectionError := net.DialTimeout("tcp", connectionTarget, connectionTimeout)
	if connectionError != nil {
		connectionCloseError := connection.Close()
		if connectionCloseError != nil {
			logger.WithError(connectionCloseError).Warning("The connection to the pinged host could not be closed")
		}
		return false
	} else {
		connectionCloseError := connection.Close()
		if connectionCloseError != nil {
			logger.WithError(connectionCloseError).Warning("The connection to the pinged host could not be closed")
		}
		return true
	}
}

// LocalIPv4Address returns the local IPv4 address which has been inferred from the network interfaces available on
// the machine. If not address could be inferred an error will be returned and the string will be empty
func LocalIPv4Address() (string, error) {
	// Get all interface addresses
	interfaceAddresses, accessErr := net.InterfaceAddrs()
	if accessErr != nil {
		return "", fmt.Errorf("unable to access the network interfaces: %w", accessErr)
	}
	// Iterate through the interface addresses...
	for _, addr := range interfaceAddresses {
		// Assert that the address is an ip address
		addr, addressIsIP := addr.(*net.IPNet)
		address := addr.IP
		if addressIsIP && !(address.IsLoopback() || address.IsMulticast()) {
			if address.To4() != nil {
				return address.String(), nil
			}
		}
	}
	return "", vars.ErrNoIPv4AddressAssigned
}
