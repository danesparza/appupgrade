package system

import (
	"net"
)

// GetOutboundIP gets the preferred outbound ip of this machine
func GetOutboundIP() (net.IP, error) {
	retval := net.IP{}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		//	todo: Handle this error by returning it
		return retval, err
	}
	defer conn.Close()

	retval = conn.LocalAddr().(*net.UDPAddr).IP

	return retval, nil
}
