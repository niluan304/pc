## pc
`pc`的全称大概是个人电脑的控制器（pc control），通过巴法云 `bemfa.com` 接入 iot 软件中, 实现远程开关机等功能。

`pc` 需要运行在一台 7*24h 的设备上，如路由器、NAS 等。

`pc` 目前只使用标准库，目的是减少依赖，也因为尚在刚开发阶段，因此很不完善，请见谅。

## Warning
由于 `pc` 采用 `go` 语言开发，因此编译文件较大 `3-5 MB`，且需要 GC，不适合在储存和内存比较小的设备运行。

`pc` 使用 `JSON` 作为配置文件，但 `JSON` 文件不支持注释，因此提供了 `JSON5` 作为配置模板，使用前，你需要修改为自己的配置，并将转化为 JSON格式，
- `JSON5` 转 `JSON`：删掉所有的注释，并将它命名为 `config.json`。
- 涉及到文件路径的部分，包括但不限于配置文件，这里都建议使用绝对路径，避免文件加载错误。

## Install
### 前置条件
1. 可以通过 `SSH` 或 `Telnet` 登录的 7*24h 的设备（后面简称为设备），如路由器、NAS
   > 开发者使用的是：小米路由器AX3000T + 原厂系统 1.0.47，仅开启了 `SSH`，没有刷机为 `OpenWrt` 系统
2. 目标主机与设备处于同一局域网内
3. 目标主机已经安装并启用 `SSH-Server` 服务，局域网内的其他设备能够通过 `SSH` 在目标主机上远程执行命令
   > 可参考教程：[Windows 上的 OpenSSH：安装、配置和使用指南 - 系统极客](https://www.sysgeek.cn/openssh-windows/)
4. 目标主机已经启用 `Wake on LAN`（需修改主板，操作系统的设置）
   > 不同主板的开启方式不同，还请读者自行搜索开启方式


### 注册巴法云账号
`pc`依赖于巴法云，因此读者需要注册一个巴法云账户，并添加一个`006`后缀的主题。
> 可参考：[平台操作教程 | 巴法文档中心](https://cloud.bemfa.com/docs/src/index_guild.html)

步骤如下：
1. 打开 [巴法物联网云平台](https://cloud.bemfa.com/)，使用邮箱或手机注册。
2. 点击「控制台」，拷贝自己的私钥，后续会用到。
3. 新建主题-命名为 `PcPower006`，名称可以为任意英文，但必须以`006`结尾，表示开关设备。

效果如图：
![bemfa init](https://github.com/niluan304/picx-images-hosting/raw/master/pc/bemfa_init.4jo078gpsw.webp)
- 点击「昵称」，网页会弹出「修改昵称」的窗口，以供自定义。

### 下载 `pc`
根据设备的芯片架构及安装的操作系统，找到对应的压缩包：[Release](https://github.com/niluan304/pc/releases/)

如果是 Linux/OpenWrt 系统，可以在命令行输入 `cat /etc/os-release | grep ARCH` 以查看设备的架构，以笔者的为例：
```bash
cat /etc/os-release | grep ARCH
LEDE_ARCH="aarch64_cortex-a53" # cortex-a53 是 实现ARMv8-A 64位指令集的微架构，故 CPU 是 arm64架构的
```

在目标设备中，将解压后的二进制文件及配置文件移动到可读写的文件夹，实现方式有两种：
- 直接通过 `curl` 或 `wget` 命令下载压缩包后解压 
- 在其他设备上下载后，通过 `scp` 等协议上传到设备

### 修改配置文件
`pc` 使用 `json`文件作为配置文件格式，但由于`json`本身不支持注释，因此示例的配置文件为 `json5` 格式，需要手动转化为 `json` 格式。
启动 `pc` 之前，使用者应当补全空白配置项的值：

```json5
// 由于标准 JSON 格式不支持注释，因此这里提供一份 JSON5的配置文件模板。
// 使用前，你需要修改为自己的配置，将转化为 JSON格式，
//
// 转化操作：删掉所有的注释，并将它命名为 config.json。
{
  // 当前程序所在主机的局域网IP，即通过 SSH/Telnet 登录的设备，一般为路由器/NAS
  // 小米路由器的局域网IP一般为：192.168.31.1
  "myIP": "",

  // 目标主机的主板网卡MAC地址，
  // windows 机器可以在命令行中输入 `ipconfig /all` 查看，如：00-1B-44-11-3A-B7
  "targetMac": "",

  // 巴法云的 UID，即控制台的私钥
  "uid": "",

  "ssh": {
    // 目标主机的 IP + SSH 端口号 
    // host:port 如 192.168.31.111:11022
    "addr": "",

    // 用户名，目前支持私钥和密码登录
    "user": "",
     
    // 通过公私钥登录，推荐使用
    // 私钥可通过 `ssh-keygen -t ed25519 -f ed25519` 命令生成
    // 然后将 ed25519 私钥上传至
    // 私钥路径，建议使用绝对地址
    "identity": "",

    // 使用密码登录，可选项
    // 密码明文，应当在局域网环境使用
    "password": ""
  },

  // topic-switch
  "switch": {
     // 主题的名称
    "topic": "XXX006",

    // Switch 只接收 on/off 两种指令，对应的操作
    // 覆盖这里的指令之前，你应该在默认的 shell，Linux(sh)/Windows(cmd) 中测试一下，以确保关机指令和取消指令是正确的。
    // Windows 命令：取消关机命令 
    "on": "cmd /c shutdown /a",
    // Windows 命令：10分钟后关机
    "off": "cmd /c shutdown /s /t 600"
  }
}
```

### 调试 `pc`

1. 启用 `pc`
   ```bash
   # Linux 机器上，赋予 `pc` 执行权限
   chmod +x pc
   # 指定配置文件并运行
   ./pc -config config.json
   ```

2. 通过巴法云推送消息

   在巴法云控制台，如果连接正常，`pc` 订阅的主题上，会显示订阅者的在线数量：
   ![](https://github.com/niluan304/picx-images-hosting/raw/master/pc/topic.8ojljeijjj.webp)

   -  设备（Windows）处于开机状态 
      - 推送 `off`，弹窗显示，即将关机：
      ![](https://github.com/niluan304/picx-images-hosting/raw/master/pc/switch-off.7zqbzd7e3r.webp)
      - 推送 `on`，弹窗显示，关机被取消：
      ![](https://github.com/niluan304/picx-images-hosting/raw/master/pc/switch-on.3d4oyo8ug8.webp)

   - 设备处于关机机状态
      在巴法云控制台，并在 `pc` 订阅的主题上推送消息。
      - 推送 `off`，设备无反应
      - 推送 `on`，设备 **开机**

3. 后台运行 `pc`
   若调试后无问题，即可在设备上后台运行 `pc`, 以捕获巴法云的消息推送
   ```bash
   ./pc -config config.json &
   ```

4. 巴法云接入 iot 软件
- 米家：![](https://github.com/niluan304/picx-images-hosting/raw/master/pc/iot-mijia.lvmqmxpuf.webp)



## RoadMap
- [x] 开关机
- [ ] 调节显示器亮度
- [ ] 开关显示器
- [ ] 使用 `Rust` 重写（低优先级）
- [ ] 其他平台（低优先级）