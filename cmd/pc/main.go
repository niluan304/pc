package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
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

	var (
		config = flag.String("config", "config.json", "config file")
	)

	flag.Parse()

	file, err := os.ReadFile(*config)
	if err != nil {
		log.Println("os.ReadFile error", err)
		panic(err)
	}

	var cfg *pc.Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		log.Println("json.Unmarshal error", err)
		panic(err)
	}

	if err := pc.Run(cfg); err != nil {
		panic(err)
	}
}
