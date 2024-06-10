package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/nilluan304/pc/bemfa"
	"github.com/nilluan304/pc/shell"
	"github.com/nilluan304/pc/wol"
)

type Config struct {
	MyIP      string `json:"myIP"`      // 当前程序所在主机的局域网IP
	TargetMac string `json:"targetMac"` // 目标主机的主板网卡MAC地址

	Uid    string `json:"uid"` // 巴法云配置
	Switch struct {
		Topic string `json:"topic"` // 巴法云 topic 名称
		On    string `json:"on"`    //  on 指令时的执行的 sh 脚本路径 / shell 脚本
		Off   string `json:"off"`   // off 指令时的执行的 sh 脚本路径 / shell 脚本
	} `json:"switch"`
}

var f = flag.String("config", "config.json", "config file")

// TODO
// handle log file
// control display by PowerShell script

func main() {
	flag.Parse()

	file, err := os.ReadFile(*f)
	if err != nil {
		log.Println("os.ReadFile error", err)
		panic(err)
	}

	var config *Config
	if err := json.Unmarshal(file, &config); err != nil {
		log.Println("json.Unmarshal error", err)
		panic(err)
	}

	s := bemfa.NewSwitch(
		func() error {
			err := wol.WakeOnLan(config.TargetMac, config.MyIP)
			if err != nil {
				return err
			}

			if _, err := shell.Shell(config.Switch.On); err != nil {
				return err
			}

			return nil
		},
		func() error {
			if _, err := shell.Shell(config.Switch.Off); err != nil {
				return err
			}
			return nil
		},
	)

	b, err := bemfa.New(config.Uid, map[string]bemfa.Topic{
		config.Switch.Topic: s,
	})
	if err != nil {
		panic(err)
	}

	b.Listen()
}
