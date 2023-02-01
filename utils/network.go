package utils

import "net"

func GetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		lg.Error("Couldn't get an IP address", err)
		return ""
	}
	for _, address := range addrs {
		// Check the address type and if it is not a loopback then display it
		ipnet, ok := address.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	lg.Error("Couldn't get an IP address", err)
	return ""
}
