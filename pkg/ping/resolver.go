package ping

import (
	"net"
)

func lookupIP(host string, ipv4, ipv6 bool) (net.IP, error) {
	ip, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	for _, i := range ip {
		if (ipv4 && isIPv4(i)) || (ipv6 && isIPv6(i)) {
			return i, nil
		}
	}
	return nil, &net.DNSError{
		Name:       host,
		Err:        "not found",
		IsNotFound: true,
	}
}

func LookupIPv4(host string) (net.IP, error) {
	return lookupIP(host, true, false)
}

func LookupIPv6(host string) (net.IP, error) {
	return lookupIP(host, false, true)
}

func LookupIP(host string) (net.IP, error) {
	return lookupIP(host, true, true)
}

var LookupFunc func(string) (net.IP, error) = LookupIP
