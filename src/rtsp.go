package sgortsp

import (
	"bufio"
	"crypto/tls"
	"errors"
	"flag"
	"log"
	"net"
	"strconv"
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

var (
	reqMethods = map[string]bool{"PLAY": true, "OPTIONS": true, "DESCRIBE": true, "PAUSE": true, "TEARDOWN": true}
	crt, key   = flag.String("crt", "server.crt", "tls crt path"), flag.String("key", "server.key", "tls key path")
	port       = flag.String("p", "9090", ":PORT")
)

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
	req, resp, err := sv.parse(conn)
	if req != nil {
		log.Println(req)
		if err != nil {
			sv.sendResponse(conn, &req.Message, err.Error())
			return
		}
		go sv.handleSession(conn, req)
		// createSessionAndListen
	} else {
		log.Println("Received a response:", resp)
	}
}

func (sv *RTSP) sendResponse(conn net.Conn, msg *Message, s string) error {
	s += CRLF
	toAppHeaders := []string{"Cseq", "Session"}
	for _, h := range toAppHeaders {
		if v, ok := msg.headers[h]; ok {
			s += h + ": " + v + CRLF
		}
	}
	_, err := conn.Write([]byte(s))
	return err
}

func (sv *RTSP) handleSession(conn net.Conn, req *Request) {
	defer conn.Close()
	session := Session{}
	transportHeader := req.headers["Transport"]
	if err := session.Init(transportHeader); err != nil {
		// response 500
		log.Println(err)
		return
	}
	req.AppendHeader("Session", session.id)
	err := sv.sendResponse(conn, &req.Message, OK)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		select {
		default:
			break
		}
		//send rtp packets using udp conn (session.conn)
	}
}

func (sv *RTSP) parse(conn net.Conn) (*Request, *Response, error) {
	var (
		req         *Request
		resp        *Response
		msg         *Message
		returnError error
	)
	s := bufio.NewScanner(conn)
	i := 0
	msgbody := false
	for s.Scan() {
		line := s.Text()
		if i == 0 && len(line) != 0 {
			// Request: method uri version
			// Response: version status-code phrase
			tokens := strings.Fields(line)
			if len(tokens) != 3 {
				// send 400 (bad request)
				log.Println("STATUS_LINE - 402 Bad request", tokens)
				returnError = errors.New(BadRequest)
			}
			method, uri, version := strings.ToUpper(tokens[0]), tokens[1], tokens[2]
			if reqMethods[method] {
				path, err := sv.Router.Parse(uri)
				if (err != nil || path == nil) && returnError == nil {
					// send 410 (gone)
					returnError = errors.New(NotFound)
				}
				if version != VERSION {
					// send 400 (bad request)
					log.Println("VERSION - 402 Bad request")
				}
				req = &Request{method: method, resource: path}
				msg = &req.Message
			} else {
				version, phrase := tokens[0], tokens[2]
				if version != VERSION {
					log.Println("VERSION - 402 Bad request")
					// return nil, nil, errors.New(VersionNotSupported)
				}
				status, err := strconv.Atoi(tokens[1])
				if err != nil && returnError == nil {
					log.Println("Status code MUST be a integer")
					returnError = errors.New(BadRequest)
				}
				log.Printf("Response: %d - %s\n", status, phrase)
				resp = &Response{status: status}
				msg = &resp.Message
			}
			i++
		} else if i != 0 {
			if !msgbody {
				sz := len(line)
				if sz == 0 {
					break
				}
				// get header's line description and assign that to headers map
				line = strings.ReplaceAll(line, "\t", "")
				line = strings.Trim(line, " ")
				idx := strings.IndexRune(line, ':')
				if idx == -1 || idx+2 >= sz {
					// invalid header, ignore it
					continue
				}
				key, value := line[:idx], line[idx+2:]
				// log.Printf("line: %s / %s -> %s\n", line, key, value)
				msg.AppendHeader(key, value)
			} else {
				msg.AppendToBody(line)
			}
			i++
		}
	}
	log.Println("headers", msg.headers)
	return req, resp, returnError
}

// RTSP dealloc
func (sv *RTSP) Close() {
}
