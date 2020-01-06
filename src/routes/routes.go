package routes

type PairString struct {
	First, Second string
}

type Resource struct {
	Path                 string
	Session, Time, Media []PairString
}

var (
	Routes = map[string]*Resource{"cade": &Resource{"/home/josec/oi.mp4",
		[]PairString{{"v", "0"}, {"o", "- %d %d IN IP4 127.0.0.1"}, {"s", "Ricky and Morty S03E01"}, {"c", "IN IP4 127.0.0.1"}, {"a", "control: *"}, {"a", "range:npt=00:00:00-00:21:00"}},
		[]PairString{{"t", "0 0"}},
		[]PairString{{"m", "video %s RTP/AVP 98"}, {"i", "Parasita Invasores alienígenas"}, {"a", "rtpmap:96 H264/90000"}},
	}}
)

// map[string]string{"v": "0", "o": "", "s": "Ricky and Morty S03E01"},
// map[string]string{"t": ""},
// map[string]string{"m": "video %s RTP/AVP 98", "i": "Parasita Invasores alienígenas", "a": "rtpmap:98 H264/90000"},
