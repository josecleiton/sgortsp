package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("converter [in] [out]")
	}
	conv(os.Args[1], os.Args[2])
}

func conv(in, out string) {
	inF, err := os.OpenFile(in, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(*inF)
	defer inF.Close()
	outF, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	twoBytes := make([]byte, 2)
	buf := make([]byte, 0, 8000)
	start := false
	n := int64(0)
	for {
		_, err := inF.Read(twoBytes)
		if err != nil {
			log.Println(err)
			break
		}
		if twoBytes[0] == 0xFF {
			if twoBytes[1] == 0xd8 {
				start = true
			} else if twoBytes[1] == 0xd9 && start {
				buf = append(buf, twoBytes...)
				num := fmt.Sprintf("%d", len(buf))
				fmt.Println(num, []byte(num))
				if len(num) <= 5 {
					s := ""
					for i := 0; i < 5-len(num); i++ {
						s += "0"
					}
					s += num
					// length := int32(len(buf))
					// header := make([]byte, 5)
					// header[1] = byte(length >> 24)
					// header[2] = byte(length >> 16)
					// header[3] = byte(length >> 8)
					// header[4] = byte(length & 0xFF)
					k, err := outF.Write([]byte(s))
					if err != nil {
						log.Fatalln(err)
					}
					n += int64(k)
					k, err = outF.Write(buf)
					if err != nil {
						log.Fatalln(err)
					}
					n += int64(k)
					log.Printf("Write bytes: %d\nHeader: %s\n", n, s)
					// log.Printf("Write bytes: %d\n", n)
				}
				buf = make([]byte, 0, 8000)
				start = false
			}
		}
		if start {
			buf = append(buf, twoBytes...)
		}
	}
	if start {
		outF.Close()
	} else {
		log.Println("nothing...")
	}
}
