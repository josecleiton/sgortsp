package main

import (
	"github.com/josecleiton/sgortsp/src"
)

func main() {
	var sv sgortsp.RTSP
	defer sv.Close()
	sv.Init()
}
