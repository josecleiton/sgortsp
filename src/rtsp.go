package sgortsp

import (
	"bufio"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/josecleiton/sgortsp/src/routes"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
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

var (
	reqMethods = map[string]bool{"PLAY": true, "OPTIONS": true, "DESCRIBE": true, "PAUSE": true, "TEARDOWN": true}
	crt, key   = flag.String("crt", "server.crt", "tls crt path"), flag.String("key", "server.key", "tls key path")
	port       = flag.String("p", "9090", ":PORT")
	serverName = flag.String("server", "RTSP Server", "server name")
	UdpPort    = flag.String("udp", "47000", "rtp packet port")
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
	// states
	caught := false
	for !caught {
		req, resp, err := sv.parse(conn)
		if req == nil {
			log.Println("Received a response:", resp)
			return
		}
		if err != nil {
			sv.sendResponse(conn, &req.Message, err.Error(), "", nil)
			return
		}
		switch req.method {
		case "SETUP":
			sv.handleSession(conn, req)
			caught = true
		case "OPTIONS":
			sv.handleOptions(conn, req)
		case "DESCRIBE":
			sv.handleDescribe(conn, req)
		default:
			log.Println("req", *req)
			log.Println(MethodNotAllowed)
		}
	}
}

func (RTSP) formatMsgBody(res *routes.Resource) string {
	if res == nil {
		return ""
	}
	var body string
	descriptors := []*[]routes.PairString{&res.Session, &res.Time, &res.Media}
	for _, descriptor := range descriptors {
		for _, desc := range *descriptor {
			switch desc.First {
			case "o":
				desc.Second = fmt.Sprintf(desc.Second, rand.Int(), rand.Int())
			case "m":
				desc.Second = fmt.Sprintf(desc.Second, *UdpPort)
			default:
				break
			}
			body += desc.First + "=" + desc.Second + CRLF
		}
	}
	if body != "" {
		body += CRLF
	}
	return body
}

func (*RTSP) sendResponse(conn net.Conn, msg *Message,
	statusLine, body string, moreHeaders map[string]string) error {

	response := statusLine + CRLF
	toAppHeaders := []string{"CSeq", "Session"}
	for _, h := range toAppHeaders {
		if v, ok := msg.headers[h]; ok {
			response += h + ": " + v + CRLF
		}
	}
	for k, v := range moreHeaders {
		response += k + ": " + v + CRLF
	}
	response += CRLF + body
	log.Printf("sendResponse:\n%s\n", response)
	_, err := conn.Write([]byte(response))
	return err
}

func (sv *RTSP) handleOptions(conn net.Conn, req *Request) error {
	headers := map[string]string{
		"Public":    "SETUP, TEARDOWN, PLAY, PAUSE, OPTIONS, DESCRIBE",
		"Supported": "play.basic",
		"Server":    *serverName,
	}
	statusLine := sv.formatStatusLine(req.version, OK)
	return sv.sendResponse(conn, &req.Message, statusLine, "", headers)
}
func (sv *RTSP) handleDescribe(conn net.Conn, req *Request) error {
	headers := map[string]string{
		"Server":       *serverName,
		"Content-Type": "application/sdp",
		"Content-Base": req.resource.Path,
	}
	statusLine := sv.formatStatusLine(req.version, OK)
	body := sv.formatMsgBody(req.resource)
	headers["Content-Length"] = strconv.Itoa(len(body))
	return sv.sendResponse(conn, &req.Message, statusLine, body, headers)
}

func (sv *RTSP) handleSession(conn net.Conn, req *Request) {
	defer conn.Close()
	session := Session{}
	if err := session.Init(req.headers["Transport"]); err != nil {
		// response 500
		log.Println(err)
		return
	}
	req.AppendHeader("Session", session.id)
	if err := sv.sendResponse(conn, &req.Message, OK, "", nil); err != nil {
		log.Println(err)
		return
	}
	play, pause, teardown := make(chan bool), make(chan bool), make(chan bool)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			req, resp, err := sv.parse(conn)
			if err != nil {
				log.Println("handleSession goro 1:", err)
				continue
			}
			if req != nil {
				switch req.method {
				case "PLAY":
					play <- true
				case "PAUSE":
					pause <- true
				case "TEARDOWN":
					teardown <- true
					return
				default:
					break
				}
			}
			if resp != nil {
				if resp.status != 200 {
					log.Println(*resp)
					teardown <- true
				}
			}
		}
	}()
	go func(send bool) {
		defer wg.Done()
	UDPLoop:
		for {
			select {
			case <-play:
				send = true
				break
			case <-pause:
				send = false
			case <-teardown:
				break UDPLoop
			default:
				break
			}
			if send {
				session.Send()
			}
		}
	}(req.method == "PLAY")
	wg.Wait()
}

func (RTSP) formatStatusLine(v string, phrase string) string {
	return v + " " + phrase
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
	version := VERSION
	for s.Scan() {
		line := s.Text()
		log.Println("line:", line)
		if i == 0 && len(line) != 0 {
			// Request: method uri version
			// Response: version status-code phrase
			tokens := strings.Fields(line)
			if len(tokens) != 3 {
				// send 400 (bad request)
				log.Println("STATUS_LINE - 402 Bad request", tokens)
				returnError = errors.New(sv.formatStatusLine(version, BadRequest))
			}
			method, uri := strings.ToUpper(tokens[0]), tokens[1]
			version = tokens[2]
			if reqMethods[method] {
				path, err := sv.Router.Parse(uri)
				log.Println("path", path)
				if (err != nil || path == nil) && returnError == nil {
					// send 410 (gone)
					returnError = errors.New(sv.formatStatusLine(version, NotFound))
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
					returnError = errors.New(sv.formatStatusLine(version, BadRequest))
				}
				log.Printf("Response: %d - %s\n", status, phrase)
				resp = &Response{status: status}
				msg = &resp.Message
			}
			msg.version = version
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
	// log.Println("headers", msg.headers)
	return req, resp, returnError
}

// RTSP dealloc
func (sv *RTSP) Close() {
}
