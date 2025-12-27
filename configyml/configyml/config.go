package configyml

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

func GetConfig(file string, data interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return DecodeConfig(f, data)
}
func DecodeConfig(r io.Reader, data interface{}) error {
	decoder := yaml.NewDecoder(r)
	return decoder.Decode(data)
}
