package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
)

var addr = flag.String("addr", "127.0.0.1:9090", "server addr")

func init() {
	flag.Parse()
}

func main() {
	log.SetFlags(log.Lshortfile)
	conf := &tls.Config{InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", *addr, conf)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	n, err := conn.Write([]byte("hello\n"))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("wrote %d bytes\n", n)
	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		fmt.Println("Client: Server public key is:")
		fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
	}
	log.Println("Client: handshake: ", state.HandshakeComplete)
	log.Println("Client: mutual: ", state.NegotiatedProtocolIsMutual)
	log.Println(state.NegotiatedProtocol)
	log.Println(state)
	// buf := make([]byte, 100)
	// n, err = conn.Read(buf)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// fmt.Println(string(buf[:n]))
}
