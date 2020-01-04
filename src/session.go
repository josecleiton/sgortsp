package sgortsp

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"net"
	"strings"
)

type Session struct {
	id string
	tr []transport
}

type transport struct {
	casttype, state int
	addrs           []*net.UDPAddr
	conns           map[int]*net.UDPConn
}

const (
	transpUnicast = iota
	transpMulticast
)

const (
	play = iota + 1
	describe
	options
	pause
	teardown
)

var (
	methodMap = map[string]int{"PLAY": play, "PAUSE": pause, "DESCRIBE": describe, "OPTIONS": options, "TEARDOWN": teardown}
)

func (s *Session) Init(transp string) error {
	var (
		idError, connError error
	)
	done := make(chan bool)
	go func() {
		idError = s.createID()
		done <- true
	}()
	go func() {
		// parse transport
		// get ip:port
		// set s.conn to that addr (use connError)
		localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1")
		if err != nil {
			log.Println(err)
		} else {
			trs := s.parseTransport(transp)
			s.connectToTransport(localAddr, &trs)
		}
		connError = err
		done <- true
	}()
	<-done
	<-done
	if idError != nil {
		return idError
	}
	return connError
}

func (s *Session) createID() error {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err != nil {
		return err
	}
	s.id = fmt.Sprintf("%x", sha256.Sum256(data))
	return nil
}

func (s *Session) parseTransport(transp string) []transport {
	transp = strings.ReplaceAll(transp, "\r\n", "")
	// MUST IMPLEMENT : MULTICAST
	requiredFields := []string{"RTP/ATP", "mode", "unicast", "dest_addr"}
	var (
		transports []transport
	)
transportLine:
	for _, str := range strings.Split(transp, ",") {
		options := make(map[string]string, 5)
		for _, tk := range strings.Split(str, ";") {
			if idx := strings.IndexRune(tk, '='); idx != -1 {
				var (
					key   string = tk[idx:]
					value string
				)
				if idx+1 < len(tk) {
					value = tk[:idx+1]
				}
				options[key] = value
			} else {
				options[tk] = ""
			}
		}
		for _, req := range requiredFields {
			if _, ok := options[req]; !ok {
				continue transportLine
			}
		}
		addrs := s.parseDestAddr(options["dest_addr"])
		state, ok := methodMap[options["mode"]]
		if !ok {
			continue transportLine
		}
		transports = append(transports, transport{transpUnicast, state, addrs, nil})
	}
	return transports
}

func (Session) connectToTransport(local *net.UDPAddr, trs *[]transport) int {
	count := 0
	for _, t := range *trs {
		if t.conns == nil {
			t.conns = make(map[int]*net.UDPConn, len(t.addrs))
		}
		for i, addr := range t.addrs {
			conn, err := net.DialUDP("udp", local, addr)
			if err != nil {
				log.Println(err)
				continue
			}
			t.conns[i] = conn
			count++
		}
	}
	return count
}

func (Session) parseDestAddr(destaddr string) []*net.UDPAddr {
	var result []*net.UDPAddr
	for _, rawAddr := range strings.Split(destaddr, "/") {
		rawAddr = strings.ReplaceAll(rawAddr, "\"", "")
		if addr, err := net.ResolveUDPAddr("udp", rawAddr); err == nil {
			result = append(result, addr)
		}
	}
	return result
}

func (s *Session) Send() {
}
