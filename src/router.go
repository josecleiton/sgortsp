package sgortsp

import (
	"errors"
	"github.com/josecleiton/sgortsp/src/routes"
	"log"
	"strings"
)

type Router struct {
	routes map[string]*routes.Resource
}

func (r *Router) Init() {
	r.routes = routes.Routes
	log.Println("r.routes:", r.routes)
}

func (r *Router) Parse(uri string) (*routes.Resource, error) {
	proute := ""
	// log.Println("uri", uri, ok)
	if ok := strings.HasPrefix(uri, "rtsp://"); !ok {
		return nil, errors.New("Bad URI")
	}
	for i, v := range uri {
		if v == '/' && i > 6 {
			proute = uri[i:]
			break
		}
	}
	if proute == "" {
		return nil, errors.New("Bad URI")
	}
	if rt := r.routes[proute]; rt != nil {
		return rt, nil
	} else {
		return rt, errors.New("Route doesn't exists")
	}
}
