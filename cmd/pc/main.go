package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"

	"gopkg.in/yaml.v3"

	"github.com/nilluan304/pc/pc"
)

var (
	version   = "v0.0.0"
	buildTime = "2006-01-02 15:04:05"
	commit    = ""
	ref       = ""
)

func main() {
	fmt.Fprintf(os.Stdout, `Welcome to PC!
Env Info:
    Git Commit: %s
    PC Version: %s
    Build Time: %s
    Go Version: %s
Ref: %s
`, commit, version, buildTime, runtime.Version(), ref)

	config := flag.String("config", "config.yml", "config file")

	flag.Parse()

	err := Load(*config)
	if err != nil {
		file, err2 := os.OpenFile("pc.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o664)
		if err2 != nil {
			panic(errors.Join(err, err2))
		}

		_, err3 := file.WriteString(err.Error() + "\n")
		panic(errors.Join(err, err2, err3))
	}
}

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %+v, error: %w", path, err)
	}

	var config *pc.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("json unmarshal: %s, error: %w", data, err)
	}

	s, err := pc.NewServer(config)
	if err != nil {
		return fmt.Errorf("new server, error: %w", err)
	}

	err = s.Run()
	if err != nil {
		return fmt.Errorf("run server, error: %w", err)
	}

	return nil
}
