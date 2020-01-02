package sgortsp

import (
	"net"
)

type Session struct {
	id   string
	conn *net.UDPAddr
}
