package wol

import (
	"bytes"
	"fmt"
	"net"
	"slices"
)

const port = 9 // WOL的默认端口号为9

// WakeOnLan
// 网络唤醒魔术包技术白皮书地址: https://www.amd.com/system/files/TechDocs/20213.pdf
func WakeOnLan(mac string, ip string) error {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return fmt.Errorf("MAC 地址错误: %s, err: %w", mac, err)
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
		return fmt.Errorf("dial udp(%s) error, err: %w", ip, err)
	}
	defer conn.Close()

	// 广播MAC地址 FF:FF:FF:FF:FF:FF
	broadcast := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	// 封装为唤醒魔术包
	magicPackage := slices.Concat(
		broadcast,
		bytes.Repeat(hw, 16),
	)
	_, err = conn.Write(magicPackage)
	if err != nil {
		return fmt.Errorf("send magic package error： %s, err: %w", magicPackage, err)
	}

	return nil
}
