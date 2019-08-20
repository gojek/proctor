package file

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func ParseYAML(filename string, procArgs map[string]string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer, err := ioutil.ReadAll(file)

	err = yaml.Unmarshal(buffer, &procArgs)
	if err != nil {
		return err
	}

	return nil
}
