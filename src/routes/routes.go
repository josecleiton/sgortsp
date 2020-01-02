package routes

type Resource struct {
	path, descriptor string
}

var Routes = map[string]Resource{"oi.mp4": {"/home/josec/oi.mp4", "v=kk"}}
