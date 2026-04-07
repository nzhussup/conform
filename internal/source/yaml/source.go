package yaml

import "github.com/nzhussup/conform/internal/schema"

type FileSource struct {
	path string
}

func NewFileSource(path string) FileSource {
	return FileSource{path: path}
}

func (s FileSource) Load(sc *schema.Schema) error {
	_ = sc
	// TODO: implement yaml loading
	return nil
}
