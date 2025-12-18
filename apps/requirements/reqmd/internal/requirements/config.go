package requirements

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/reqmd/internal/helper"
)

type Config struct {
	Aspects map[string][]string `json:"Aspects"` // Ordered list of aspects, any value okay for them.
}

func ParseConfig(filename string) (config Config, err error) {

	contents, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, errors.WithStack(err)
	}

	if err = json.Unmarshal(contents, &config); err != nil {
		return Config{}, errors.WithStack(err)
	}

	return config, nil
}

func (c *Config) String() (value string) {
	return helper.JsonPretty(c)
}
