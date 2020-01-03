package routes

type Resource struct {
	Path, Descriptor string
}

var Routes = map[string]*Resource{"oi.mp4": &Resource{"/home/josec/oi.mp4", "v=kk"}}
