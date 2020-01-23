package routes

type PairString struct {
	First, Second string
}

type Resource struct {
	Path                 string
	Session, Time, Media []PairString
}

var (
	Routes = map[string]*Resource{"cade": &Resource{"/home/josec/out.mjpg",
		[]PairString{{"v", "0"}},
		[]PairString{},
		[]PairString{{"m", "video %s RTP/AVP 26"}, {"a", "a=control:streamid=1234"}},
	}}
)

// , {"o", "- %d %d IN IP4 127.0.0.1"}, {"s", "Ricky and Morty S03E01"}, {"c", "IN IP4 127.0.0.1"}, {"a", "control: *"}
// {"t", "0 0"}
//
// {"a", "range:npt=00:00:00-00:21:00"}
// {"i", "Parasita Invasores alienígenas"}

// map[string]string{"v": "0", "o": "", "s": "Ricky and Morty S03E01"},
// map[string]string{"t": ""},
// map[string]string{"m": "video %s RTP/AVP 98", "i": "Parasita Invasores alienígenas", "a": "rtpmap:98 H264/90000"},
