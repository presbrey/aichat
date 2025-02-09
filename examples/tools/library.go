package tools

import (
	_ "embed"

	"github.com/presbrey/aichat"
	"gopkg.in/yaml.v3"
)

//go:embed library.yaml
var yamlBytes []byte

var Library = map[string]map[string]*aichat.Tool{}

func init() {
	yaml.Unmarshal(yamlBytes, &Library)
}

func Get(key string) []*aichat.Tool {
	m := Library[key]
	if m == nil {
		return nil
	}
	values := make([]*aichat.Tool, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
