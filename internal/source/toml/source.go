package toml

import (
	"github.com/nzhussup/konform/internal/schema"
	"github.com/nzhussup/konform/internal/source/common"
	"github.com/pelletier/go-toml/v2"
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
	return common.LoadFileWithMode(sc, s.path, s.callerDir, "toml", s.unmarshalDocument, s.suggestionMode)
}

func (s Source) LoadBytes(sc *schema.Schema) error {
	return common.LoadBytesWithMode(sc, s.data, "toml", s.unmarshalDocument, s.suggestionMode, "toml")
}

func (s Source) unmarshalDocument(data []byte) (common.Document, error) {
	var doc common.Document
	if err := toml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}
