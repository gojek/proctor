package file

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"proctor/internal/pkg/model/metadata/env"
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

func WriteYAML(filename string, procArgs []env.VarMetadata) error {
	var content string

	for _, procArg := range procArgs {
		content += "# " + procArg.Description
		content += "\n"
		content += procArg.Name + ":"
		content += "\n"
	}

	err := ioutil.WriteFile(filename, []byte(content), 0644)
	return err
}
