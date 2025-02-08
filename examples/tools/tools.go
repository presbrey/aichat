package tools

import (
	_ "embed"

	"github.com/presbrey/aichat"
	"gopkg.in/yaml.v3"
)

//go:embed tools.yaml
var yamlBytes []byte

var Library = map[string][]*aichat.Tool{}

func init() {
	yaml.Unmarshal(yamlBytes, &Library)
}
