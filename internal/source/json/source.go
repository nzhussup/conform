package json

import (
	"encoding/json"

	"github.com/nzhussup/konform/internal/schema"
	"github.com/nzhussup/konform/internal/source/common"
)

type FileSource struct {
	path      string
	callerDir string
}

func NewFileSource(path string, callerDir string) FileSource {
	return FileSource{path: path, callerDir: callerDir}
}

func (s FileSource) Load(sc *schema.Schema) error {
	return common.LoadFile(sc, s.path, s.callerDir, "json", func(data []byte) (common.Document, error) {
		var doc common.Document
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, err
		}
		return doc, nil
	})
}
