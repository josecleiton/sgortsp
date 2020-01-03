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

const (
	BADREQUEST = 402
)

var reqMethods = map[string]bool{"PLAY": true, "OPTIONS": true, "DESCRIBE": true, "PAUSE": true, "TEARDOWN": true}
var crt, key = flag.String("crt", "server.crt", "tls crt path"), flag.String("key", "server.key", "tls key path")
var port = flag.String("p", "9090", ":PORT")

func init() {
	flag.Parse()
}

// RTSP Initializer
func (sv *RTSP) Init() {
	sv.sessions = make(map[string]Session, 10)
	sv.Router.Init()
	sv.listen()
}

func (sv *RTSP) listen() {
	port := *port
	// config := sv.setupTLS()
	// ln, err := tls.Listen("tcp", ":"+port, config)
	ln, err := net.Listen("tcp", ":"+port)
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
	desc, err := sv.parse(conn)
	if err != nil {
		return
	}
	if desc == REQUEST {
	} else {
		// go sv.handleConn(conn)
	}
	desc++
}

func (sv *RTSP) handleConn(conn net.Conn) {
	for {

	}
}

func (sv *RTSP) parse(conn net.Conn) (int, error) {
	s := bufio.NewScanner(conn)
	i := 0
	msgtype := UNDEFINED
	msgbody := false
	body := ""
	headers := make(map[string]string, 5)
	for s.Scan() {
		line := s.Text()
		if i == 0 && len(line) != 0 {
			// Request: method uri version
			// Response: version status-code phrase
			tokens := strings.Fields(line)
			if len(tokens) != 3 {
				// send 400 (bad request)
				log.Println("STATUS_LINE - 402 Bad request", tokens)
			}
			method, uri, version := strings.ToUpper(tokens[0]), tokens[1], tokens[2]
			if reqMethods[method] {
				msgtype = REQUEST
				path, err := sv.Router.Parse(uri)
				if err != nil || path == nil {
					// send 410 (gone)
					log.Println("410 GONE")
				}
				if version != VERSION {
					// send 400 (bad request)
					log.Println("VERSION - 402 Bad request")
				}
			} else {
				msgtype = RESPONSE
			}
			i++
		} else if i != 0 {
			log.Println(line, "WOW")
			if !msgbody {
				sz := len(line)
				if sz == 0 {
					break
				}
				// get header's line description and assign that to headers map
				idx := strings.IndexRune(line, ':')
				if idx == -1 || idx+2 >= sz {
					// invalid header, ignore it
					continue
				}
				key, value := line[:idx], line[idx+2:]
				// log.Printf("line: %s / %s -> %s\n", line, key, value)
				headers[key] = strings.Trim(value, " ")
			} else {
				log.Println("body", body)
				body += line
			}
			i++
		}
	}
	log.Println("headers", headers)
	return msgtype, nil
}
func (sv *RTSP) sendError() {
}

// RTSP dealloc
func (sv *RTSP) Close() {
}
