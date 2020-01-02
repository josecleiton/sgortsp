package sgortsp

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"os"
	"strings"
)

const (
	CRLF    = "\r\n"
	VERSION = "RTSP/2.0"
)

const (
	UNDEFINED = iota
	REQUEST
	RESPONSE
)

var reqMethods = map[string]bool{"PLAY": true}

// RTSP main type
// Manage Requests/Responses and Sessions
type RTSP struct {
	sessions map[string]Session
}

// RTSP Initializer
func (sv *RTSP) Init() {
	sv.sessions = make(map[string]Session, 10)
	sv.listen()
}

func (sv *RTSP) listen() {
	port := sv.setupPort()
	config := sv.setupTLS()
	ln, err := tls.Listen("tcp", port, config)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("RTSP Server running on %s!\n", port)
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go sv.handleConn(conn)
	}
}

func (sv *RTSP) setupPort() string {
	port := ":9090"
	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}
	return port
}

func (sv *RTSP) setupTLS() *tls.Config {
	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalln(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{cer}}
}

func (sv *RTSP) handleConn(conn net.Conn) {
	sv.parser(conn)
}

func (sv *RTSP) parser(conn net.Conn) int {
	s := bufio.NewScanner(conn)
	i := -1
	msgtype := UNDEFINED
	msgbody := false
	body := ""
	headers := make(map[string]string, 5)
	for s.Scan() {
		line := s.Text()
		if i == -1 && len(line) != 0 {
			i++
		} else if i != 0 {
			if !msgbody {
				// get header's line description and assign that to headers map
				idx := strings.IndexRune(line, ':')
				if idx == -1 {
					continue
				}
				key, value := line[:idx], line[idx+1:]
				headers[key] = strings.Trim(value, " ")
			} else {
				body += line
			}
			i++
		} else {
			tokens := strings.Fields(line)
			tokens[0] = strings.ToUpper(tokens[0])
			if len(tokens) != 3 {
				// send 400 (bad request)
			}
			if reqMethods[tokens[0]] {
				msgtype = REQUEST
				if tokens[2] != VERSION {
					// send 400 (bad request)
				}
			} else {
				msgtype = RESPONSE
			}
			i++
		}
	}
	return msgtype
}

// RTSP dealloc
func (sv *RTSP) Close() {
}
