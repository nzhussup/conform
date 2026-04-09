package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/nzhussup/konform/internal/schema"
	"github.com/nzhussup/konform/internal/source/common"
)

type Source struct {
	path           string
	callerDir      string
	data           []byte
	suggestionMode common.UnknownKeySuggestionMode
}

func NewFileSource(path string, callerDir string, suggestionMode common.UnknownKeySuggestionMode) Source {
	return Source{path: path, callerDir: callerDir, suggestionMode: suggestionMode}
}

func NewByteSource(data []byte, suggestionMode common.UnknownKeySuggestionMode) Source {
	return Source{data: data, suggestionMode: suggestionMode}
}

func (s Source) LoadFile(sc *schema.Schema) error {
	return common.LoadFileWithMode(sc, s.path, s.callerDir, "yaml", s.unmarshalDocument, s.suggestionMode)
}

func (s Source) LoadBytes(sc *schema.Schema) error {
	return common.LoadBytesWithMode(sc, s.data, "yaml", s.unmarshalDocument, s.suggestionMode, "yaml")
}

func (s Source) unmarshalDocument(data []byte) (common.Document, error) {
	var doc common.Document
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}
