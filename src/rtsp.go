package sgortsp

import (
	"bufio"
	"crypto/tls"
	"flag"
	"log"
	"net"
	"strings"
)

// RTSP main type
// Manage Requests/Responses and Sessions
type RTSP struct {
	sessions map[string]Session
	Router
}

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
var crt, key = flag.String("crt", "server.crt", "tls crt path"), flag.String("key", "server.key", "tls key path")
var port = flag.String("p", "9090", ":PORT")

func init() {
	flag.Parse()
}

// RTSP Initializer
func (sv *RTSP) Init() {
	sv.sessions = make(map[string]Session, 10)
	sv.listen()
}

func (sv *RTSP) listen() {
	port := *port
	config := sv.setupTLS()
	ln, err := tls.Listen("tcp", ":"+port, config)
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
		go sv.handShake(conn)
		// go func() {
		// 	s := bufio.NewScanner(conn)
		// 	for s.Scan() {
		// 		log.Println(s.Text())
		// 	}
		// }()
	}
}

func (sv *RTSP) setupTLS() *tls.Config {
	cer, err := tls.LoadX509KeyPair(*crt, *key)
	if err != nil {
		log.Fatalln(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{cer}}
}

func (sv *RTSP) handShake(conn net.Conn) {
	sv.parse(bufio.NewScanner(conn))
}

func (sv *RTSP) handleConn(conn net.Conn) {
}

func (sv *RTSP) parse(s *bufio.Scanner) int {
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
				path, err := sv.Router.Parse(tokens[1])
				if err != nil {
					// send 410 (gone)
				}
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
