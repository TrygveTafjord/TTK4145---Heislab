package network

import (
	"fmt"
	"net"
	"strings"
)

var localIP string

func LocalIP(port string) (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]

		localIP = fmt.Sprintf("peer-%s-%d", localIP, port)
	}
	return localIP, nil
}
