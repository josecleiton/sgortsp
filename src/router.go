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
	if ok := strings.HasPrefix(uri, "rtsp://"); !ok {
		return nil, errors.New("Bad URI")
	}
	for i, v := range uri {
		if v == '/' && i > 6 {
			proute = uri[i+1:]
			break
		}
	}
	if proute == "" {
		proute = "/"
	} else {
		proute = strings.TrimSuffix(proute, "/")
	}
	if rt := r.routes[proute]; rt != nil {
		return rt, nil
	}
	return nil, errors.New("Route doesn't exists")
}
