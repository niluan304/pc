package pc

import (
	"errors"
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

	if s.Identity != "" {
		authMethod, _ := WithPrivate(s.Identity)
		if authMethod != nil {
			auth = append(auth, authMethod)
		}
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
		return nil, errors.Join(err, fmt.Errorf("connect to server (%s) failed", s.Addr))
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("create session failed"))
	}

	defer session.Close()

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("execute command (%s) failed", cmd))
	}
	return out, nil
}

func WithPrivate(private string) (ssh.AuthMethod, error) {
	pem, err := os.ReadFile(private)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("read private key file (%s) failed", private))
	}

	signer, err := ssh.ParsePrivateKey(pem)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("parse private key file (%s) failed", private))
	}

	return ssh.PublicKeys(signer), nil
}
