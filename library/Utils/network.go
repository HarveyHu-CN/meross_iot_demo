package Utils

import "net"

func GetOneLocalIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "127.0.0.1", nil
}

func GetOneLocalMac() (string, error) {
	ifcs, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, ifc := range ifcs {
		mac := ifc.HardwareAddr
		macStr := mac.String()
		if macStr != "" {
			return macStr, nil
		}
	}
	return "", nil
}
