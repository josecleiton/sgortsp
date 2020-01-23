package sgortsp

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type Session struct {
	id, ip string
	tr     []transport
	File   Streamer
}

type transport struct {
	casttype, state int
	ports           []string
	conn            *net.UDPConn
	Rtp             RTP
	// addrs           []*net.UDPAddr
}

const (
	transpUnicast = iota
	transpMulticast
)

const (
	ServerPort = "40400"
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

func (s *Session) Init(a net.Addr, transp string) error {
	var (
		idError, connError error
	)
	remoteAddr, ok := a.(*net.TCPAddr)
	if !ok {
		return errors.New("Session init: malformed remoteaddr")
	}
	s.ip = remoteAddr.IP.String()
	log.Println("session ip:", s.ip)
	done := make(chan bool)
	go func() {
		idError = s.createID()
		done <- true
	}()
	go func() {
		// parse transport
		// get ip:port
		// set s.conn to that addr (use connError)
		localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+ServerPort)
		if err != nil {
			log.Println(err)
		} else {
			s.tr = s.ParseTransport(transp)
			s.connectToTransport(localAddr, &s.tr)
			log.Println("session:", *s)
		}
		connError = err
		done <- true
	}()
	<-done
	<-done
	if len(s.tr) == 0 {
		return errors.New("None transport available")
	}
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

func (s *Session) ParseTransport(transp string) []transport {
	transp = strings.ReplaceAll(transp, "\r\n", "")
	// MUST IMPLEMENT : MULTICAST
	requiredFields := []string{"unicast", "client_port"}
	var (
		transports []transport
	)
transportLine:
	for _, str := range strings.Split(transp, ",") {
		options := make(map[string]string, 5)
		for _, tk := range strings.Split(str, ";") {
			tk = strings.Trim(tk, " ")
			if idx := strings.IndexRune(tk, '='); idx != -1 {
				key, value := tk[:idx], tk[idx+1:]
				options[key] = value
				log.Printf("%s-%s\n", key, value)
			} else {
				options[tk] = ""
			}
		}
		types := []string{"RTP/AVP", "RTP/AVP/UDP", "RTP/UDP"}
		exists := false
		for _, t := range types {
			if _, ok := options[t]; ok {
				exists = true
			}
		}
		if !exists {
			continue transportLine
		}
		for _, req := range requiredFields {
			if _, ok := options[req]; !ok {
				log.Println("FAILED", req, options[req])
				continue transportLine
			}
		}
		ports := s.parseClientPort(options["client_port"])
		log.Println("ports", ports)
		state, ok := methodMap[strings.ToUpper(options["mode"])]
		if !ok {
			state = methodMap["PLAY"]
		}
		transports = append(transports, transport{
			casttype: transpUnicast,
			state:    state,
			ports:    ports,
			conn:     nil,
		})
	}
	return transports
}

func (s *Session) connectToTransport(local *net.UDPAddr, trs *[]transport) int {
	count := 0
	log.Println("trans", *trs)
	for i, t := range *trs {
		for _, port := range t.ports {
			remote, err := net.ResolveUDPAddr("udp", s.ip+":"+port)
			if err != nil {
				log.Println("connectToTransport:", err)
				count++
				continue
			}
			conn, err := net.DialUDP("udp", local, remote)
			if err == nil {
				(*trs)[i].conn = conn
				break
			}
			log.Println(err)
			count++
		}
	}
	// log.Println(trs)
	return count
}

func (Session) parseClientPort(clientport string) []string {
	ports := make([]string, 0, 2)
	if idx := strings.IndexRune(clientport, '-'); idx != -1 {
		ports = append(ports, clientport[:idx], clientport[idx+1:])
	} else {
		ports = append(ports, clientport)
	}
	log.Println("ports", ports)
	return ports
}

func (s *Session) parseDestAddr(destaddr string) []*net.UDPAddr {
	var result []*net.UDPAddr
	for _, rawPort := range strings.Split(destaddr, "-") {
		addr, err := net.ResolveUDPAddr("udp", s.ip+":"+rawPort)
		if err != nil {
			log.Println("parseDestAddr:", err)
			continue
		}
		result = append(result, addr)
	}
	return result
}

func (s *Session) Send() error {
	// log.Println(s.tr)
	if s.File.FrameN > 0 {
		time.Sleep(50 * time.Millisecond)
	}
	if err := s.File.nextFrame(); err != nil {
		log.Println(err)
		return err
	}
	payloadType, frameN, framePeriod := s.File.Type, s.File.FrameN, s.File.FramePeriod
	for _, tr := range s.tr {
		packet := tr.Rtp.Packet(s.File.Data, payloadType, frameN, framePeriod)
		_, err := tr.conn.Write(packet)
		if err != nil {
			log.Println(err)
		}
		// log.Println("packet", n, len(packet))
	}
	return nil
}
