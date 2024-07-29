package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"

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

	config := flag.String("config", "config.json", "config file")

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

func Load(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read file: %+v, error: %w", file, err)
	}

	var config *pc.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("json unmarshal: %s, error: %w", data, err)
	}

	err = pc.Run(config)
	if err != nil {
		return fmt.Errorf("pc run, error: %w", err)
	}
	return nil
}
