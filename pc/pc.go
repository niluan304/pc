package pc

import (
	"github.com/nilluan304/pc/bemfa"
	"github.com/nilluan304/pc/wol"
)

type Config struct {
	MyIP      string `json:"myIP"`      // 当前程序所在主机的局域网IP
	TargetMac string `json:"targetMac"` // 目标主机的主板网卡MAC地址

	SSH *SSH `json:"ssh,omitempty"` // 目标主机的 ssh 配置

	Uid    string `json:"uid"` // 巴法云配置
	Switch struct {
		Topic string `json:"topic"` // 巴法云 topic 名称
		On    string `json:"on"`    //  on 指令时的执行的 shell 脚本
		Off   string `json:"off"`   // off 指令时的执行的 shell 脚本
	} `json:"switch"`
}

func Run(config *Config) error {
	s := bemfa.NewSwitch(
		func() error {
			_ = config.SSH.Command(config.Switch.On)

			if err := wol.WakeOnLan(config.TargetMac, config.MyIP); err != nil {
				return err
			}

			return nil
		},
		func() error {
			if err := config.SSH.Command(config.Switch.Off); err != nil {
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

	return b.Listen()
}
