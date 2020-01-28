package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var c *bool = flag.Bool("c", false, "convert only")
var in *string = flag.String("i", "", "input video")
var out *string = flag.String("o", "", "output mjpg video")

func init() {
	flag.Parse()
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("converter -i video -o o.mjpg")
	}
	src := *in
	log.Println(*c, *in, *out)
	if !*c {
		src = ffmpeg(src)
		defer os.Remove(src)
	}
	conv(src, *out)
}

func conv(in, out string) {
	inF, err := os.OpenFile(in, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer inF.Close()
	outF, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0644)
	defer outF.Close()
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
			fmt.Print("\n")
			log.Println(err)
			break
		}
		if twoBytes[0] == 0xFF {
			if twoBytes[1] == 0xd8 {
				start = true
			} else if twoBytes[1] == 0xd9 && start {
				buf = append(buf, twoBytes...)
				num := fmt.Sprintf("%d", len(buf))
				// fmt.Println(num, []byte(num))
				if len(num) <= 5 {
					s := prefixWithZeroes(num, 5)
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
					fmt.Printf("\rbytes: %d", n)
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
}

func prefixWithZeroes(s string, n int) string {
	ans := ""
	for i := 0; i < n-len(s); i++ {
		ans += "0"
	}
	ans += s
	return ans
}

func ffmpeg(src string) string {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "rtsp-*.mjpg")
	if err != nil {
		log.Fatalln("cannot create tmp file", err)
	}
	defer tmpFile.Close()
	fName := tmpFile.Name()
	log.Println("name:", fName)
	cmd := exec.Command("ffmpeg", "-i", src, fName)
	cmd.Stdout = os.Stdout
	// cmd := exec.Command("echo", "323131")
	if err := cmd.Run(); err != nil {
		log.Fatalln("ffmpeg failed", err)
	}
	return fName
}
