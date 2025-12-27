package argparse

import (
	"os"
	"strings"
)

var arguments map[string]string
var initialized bool

func InitArgs() {
	initialized = true
	arguments = make(map[string]string)
	for _, arg := range os.Args {
		parts := strings.Split(arg, "=")
		if len(parts) > 1 {
			if strings.Index(parts[0], "-") == 0 {
				value := strings.Join(parts[1:], "=")
				key := strings.ToLower(parts[0][1:])

				arguments[key] = value
			}
		}
	}
}
func GetArg(key string, defaultval string) string {
	if !initialized {
		InitArgs()
	}
	value := arguments[strings.ToLower(key)]
	if value == "" {
		return defaultval
	}
	return value
}
