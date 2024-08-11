package pc

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSH struct {
	Addr     string `json:"addr"`
	User     string `json:"user"`
	Identity string `json:"identity"`
	Password string `json:"password"`
}

func (s *SSH) Command(cmd string) ([]byte, error) {
	var auth []ssh.AuthMethod
	if s.Password != "" {
		auth = append(auth, ssh.Password(s.Password))
	}

	if authMethod, _ := WithPrivate(s.Identity); authMethod != nil {
		auth = append(auth, authMethod)
	}

	client, err := ssh.Dial("tcp", s.Addr, &ssh.ClientConfig{
		User:              s.User,
		Auth:              auth,
		HostKeyCallback:   ssh.InsecureIgnoreHostKey(), // ignore know_host
		BannerCallback:    ssh.BannerDisplayStderr(),
		ClientVersion:     "",
		HostKeyAlgorithms: nil,
		Timeout:           20 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to server (%s) failed, err: %w", s.Addr, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("create session failed, err: %w", err)
	}

	defer session.Close()

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, fmt.Errorf("execute command (%s) failed, err: %w", cmd, err)
	}
	return out, nil
}

func WithPrivate(private string) (ssh.AuthMethod, error) {
	pem, err := os.ReadFile(private)
	if err != nil {
		return nil, fmt.Errorf("read private key file (%s) failed, err: %w", private, err)
	}

	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		return nil, fmt.Errorf("parse private key file (%s) failed, err: %w", private, err)
	}

	return ssh.PublicKeys(signer), nil
}
