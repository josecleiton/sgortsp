package sgortsp

import (
	"errors"
	"github.com/josecleiton/sgortsp/src/routes"
	"strings"
)

type Router struct {
	routes map[string]routes.Resource
}

func (r *Router) Init() {
	r.routes = routes.Routes
}

func (r *Router) Parse(uri string) (*routes.Resource, error) {
	ok := strings.HasPrefix(uri, "rtsp://")
	var proute string
	if !ok {
		return nil, errors.New("KK")
	}
	for i, v := range uri {
		if v == '/' && i > 6 {
			proute = uri[i:]
			break
		}
	}
	if proute == "" {
		return nil, errors.New("")
	}
	rt := r.routes[proute]
	return &rt, nil
}
