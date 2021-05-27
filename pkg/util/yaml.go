package util

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v2"
)

// SplitYAML is used to split yaml
func SplitYAML(data []byte) ([][]byte, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var parts [][]byte
	for {
		var value interface{}
		err := decoder.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		part, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, nil
}

// DumpYaml is used to dump yaml into stdout
func DumpYaml(cfg interface{}) {
	out, err := yaml.Marshal(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Printf("%s\n", out)
	}
}

// LoadConfig read YAML-formatted config from filename into cfg.
func LoadConfig(filename string, pointer interface{}) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return multierror.Prefix(err, "Error reading config file")
	}

	err = yaml.UnmarshalStrict(buf, pointer)
	if err != nil {
		return multierror.Prefix(err, "Error parsing config file")
	}

	return nil
}
