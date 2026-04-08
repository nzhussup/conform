package toml

import (
	"github.com/nzhussup/konform/internal/schema"
	"github.com/nzhussup/konform/internal/source/common"
	"github.com/pelletier/go-toml/v2"
)

type FileSource struct {
	path      string
	callerDir string
}

func NewFileSource(path string, callerDir string) FileSource {
	return FileSource{path: path, callerDir: callerDir}
}

func (s FileSource) Load(sc *schema.Schema) error {
	return common.LoadFile(sc, s.path, s.callerDir, "toml", func(data []byte) (common.Document, error) {
		var doc common.Document
		if err := toml.Unmarshal(data, &doc); err != nil {
			return nil, err
		}
		return doc, nil
	})
}
