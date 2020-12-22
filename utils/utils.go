package utils

import (
	"fmt"
	"log"
	"math/rand"
	"net"
)

// GetOutboundIP returns the preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

//GenRandomHexCode generates a random 6-digit hex color code.
func GenRandomHexCode() string {
	var code string
	for i := 0; i < 3; i++ {
		code += fmt.Sprintf("%02X", rand.Intn(255))
	}
	fmt.Println(code)
	return code
}
