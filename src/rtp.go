package sgortsp

import (
	"log"
	"math/rand"
)

type RTP struct {
	ssrc uint32
	init bool
}

const (
	version   = 2
	padding   = 0
	extension = 0
	cc        = 0
	marker    = 0
	headerLen = 12
)

func (r *RTP) genSSRC() {
	r.ssrc = rand.Uint32()
}

func (r *RTP) Packet(body []byte, payloadType byte, frameN, framePeriod int) []byte {
	if !r.init {
		r.ssrc = 2
		r.init = true
	}
	// log.Println(r.ssrc)
	timestamp := int32(frameN * framePeriod)
	sequenceNumber := int16(frameN)
	header := []byte{
		version<<6 | padding<<5 | extension<<4 | cc,
		marker<<7 | payloadType,
		byte(sequenceNumber >> 8),
		byte(sequenceNumber & 0xFF),
		byte(timestamp >> 24),
		byte(timestamp >> 16),
		byte(timestamp >> 8),
		byte(timestamp & 0xFF),
		byte(r.ssrc >> 24),
		byte(r.ssrc >> 16),
		byte(r.ssrc >> 8),
		byte(r.ssrc & 0xFF),
	}
	payload := make([]byte, 0, len(header)+len(body)+10)
	payload = append(payload, header...)
	log.Println("header", payload)
	return append(payload, body...)
}
