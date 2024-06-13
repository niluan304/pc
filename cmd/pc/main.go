package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/nilluan304/pc/pc"
)

var config = flag.String("config", "config.json", "config file")

// TODO
// handle log file
// control display by PowerShell script

func main() {
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
