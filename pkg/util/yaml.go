package util

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v2"
)

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
