package sgortsp

import (
	"github.com/josecleiton/sgortsp/src/routes"
)

const (
	// status code
	OK                   = "200 OK"
	BadRequest           = "400 Bad Request"
	Unauthorized         = "401 Unauthorized"
	Forbidden            = "403 Forbidden"
	NotFound             = "404 Not Found"
	MethodNotAllowed     = "405 Method Not Allowed"
	NotAcceptable        = "406 Not Acceptable"
	Gone                 = "410 Gone"
	RequestTooLarge      = "413 Request Message Body Too Large"
	UnsupportedMedia     = "415 Unsupported Media Type"
	SessionNotFound      = "453 Session Not Found"
	NotValidState        = "455 Method Not Valid in This State"
	HeaderNotValid       = "456 Header Field not Valid for Resource"
	InvalidRange         = "457 Invalid Range"
	UnsupportedTransport = "461 Unsupported Transport"
	InternalError        = "500 Internal Server Error"
	NotImplemented       = "501 Not Implemented"
	BadGateway           = "502 Bad Gateway"
	Unavailable          = "503 Service Unavailable"
	VersionNotSupported  = "505 RTSP Version Not Supported"
	OptionNotSupported   = "551 Option Not Supported"
)

type Message struct {
	body, version string
	headers       map[string]string
}

type Msg interface {
	AppendHeader(string, string)
	AppendToBody(string)
}

type Request struct {
	Message
	method   string
	resource *routes.Resource
}

type Response struct {
	Message
	status int
}

func (m *Message) AppendHeader(k, v string) {
	if m.headers == nil {
		m.headers = make(map[string]string, 5)
	}
	m.headers[k] = v
}

func (m *Message) AppendToBody(s string) {
	m.body += s
}
