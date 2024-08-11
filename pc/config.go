package pc

type Config struct {
	MyIP      string `yaml:"myIP"`      // 当前程序所在主机的局域网IP
	TargetMac string `yaml:"targetMac"` // 目标主机的主板网卡MAC地址

	SSH *SSH `yaml:"ssh,omitempty"` // 目标主机的 ssh 配置

	Log *Log `yaml:"log,omitempty"` // 日志配置

	Bemfa *Bemfa `yaml:"bemfa"` // 巴法云配置
}

type Bemfa struct {
	Uid    string `yaml:"uid"` // 巴法云配置
	Switch struct {
		Topic string `yaml:"topic"` // 巴法云 topic 名称
		On    string `yaml:"on"`    //  on 指令时的执行的 shell 脚本
		Off   string `yaml:"off"`   // off 指令时的执行的 shell 脚本
	} `yaml:"switch"`
}
