package sgortsp

import (
	"os"
	"strconv"
)

const (
	mjpegPeriod = 100
	mjpegCode   = 26
)

type Streamer struct {
	FrameN      int
	FramePeriod int
	Type        byte
	Data        []byte
	file        *os.File
	frameLen    []byte
}

func (s *Streamer) Init(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	s.file = file
	s.FramePeriod = mjpegPeriod
	s.Type = mjpegCode
	s.frameLen = make([]byte, 5)
	return nil
}

func (s *Streamer) NextFrame() error {
	if _, err := s.file.Read(s.frameLen); err != nil {
		return err
	}
	// frame := s.frameLen
	// length := int(frame[0]<<24) | int(frame[1]<<16) |
	// 	int(frame[2]<<8) | int(frame[3]&0xFF)
	length, err := strconv.Atoi(string(s.frameLen))
	if err != nil {
		return err
	}
	s.FrameN++
	s.Data = make([]byte, length)
	s.file.Read(s.Data)
	return nil
}

func (s *Streamer) Close() {
	s.file.Close()
}
