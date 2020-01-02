package main

import (
	"crypto/tls"
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
	conf := &tls.Config{}
	conn, err := tls.Dial("tcp", *addr, conf)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	n, err := conn.Write([]byte("hello\n"))
	if err != nil {
		log.Fatalln(err)
	}
	buf := make([]byte, 100)
	n, err = conn.Read(buf)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(buf[:n]))
}
