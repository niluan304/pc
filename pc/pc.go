package pc

import (
	"github.com/nilluan304/pc/bemfa"
	"github.com/nilluan304/pc/wol"
)

type Config struct {
	MyIP      string `json:"myIP"`      // 当前程序所在主机的局域网IP
	TargetMac string `json:"targetMac"` // 目标主机的主板网卡MAC地址

	SSH *SSH `json:"ssh,omitempty"` // 目标主机的 ssh 配置

	Log *Log `json:"log,omitempty"` // 日志配置

	Uid    string `json:"uid"` // 巴法云配置
	Switch struct {
		Topic string `json:"topic"` // 巴法云 topic 名称
		On    string `json:"on"`    //  on 指令时的执行的 shell 脚本
		Off   string `json:"off"`   // off 指令时的执行的 shell 脚本
	} `json:"switch"`
}

func Run(config *Config) error {
	s := bemfa.NewSwitch(
		// todo handle Command out

		func() (err error) {
			out, _ := config.SSH.Command(config.Switch.On)
			_ = out

			err = wol.WakeOnLan(config.TargetMac, config.MyIP)
			if err != nil {
				return err
			}

			return nil
		},
		func() (err error) {
			out, err := config.SSH.Command(config.Switch.Off)
			if err != nil {
				return err
			}

			_ = out

			return nil
		},
	)

	logger, err := config.Log.Logger()
	if err != nil {
		return err
	}

	topics := map[string]bemfa.Topic{
		config.Switch.Topic: s,
	}

	b, err := bemfa.New(config.Uid, topics, bemfa.WithLogger(logger.WithGroup("bemfa")))
	if err != nil {
		return err
	}

	logger.Info("pc start", "listen bemfa", topics)

	return b.Listen()
}
