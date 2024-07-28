package wol

import (
	"bytes"
	"errors"
	"fmt"
	"net"
)

const port = 9 // WOL的默认端口号为9

// WakeOnLan
// 网络唤醒魔术包技术白皮书地址: https://www.amd.com/system/files/TechDocs/20213.pdf
func WakeOnLan(mac string, ip string) error {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return errors.Join(err, fmt.Errorf("MAC: 输入不正确: %s", mac))
	}

	// 广播MAC地址 FF:FF:FF:FF:FF:FF
	broadcast := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	buf := bytes.NewBuffer(broadcast)

	for range 16 {
		buf.Write(hw)
	}

	sender := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
		Zone: "",
	}
	target := net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 0,
		Zone: "",
	}
	conn, err := net.DialUDP("udp", &sender, &target)
	if err != nil {
		return errors.Join(err, fmt.Errorf("dial udp(%s) error", ip))
	}
	defer conn.Close()

	// 获得唤醒魔术包
	mp := buf.Bytes()
	_, err = conn.Write(mp)
	if err != nil {
		return errors.Join(err, fmt.Errorf("send magic package error： %s", mp))
	}

	return nil
}
