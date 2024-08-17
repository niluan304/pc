package pc

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/nilluan304/pc/bemfa"
	"github.com/nilluan304/pc/wol"
)

type Server struct {
	logger *slog.Logger
	bemfa  *bemfa.Bemfa
}

func NewServer(config *Config) (*Server, error) {
	// todo check config

	logger, err := config.Log.Logger()
	if err != nil {
		return nil, err
	}

	topics := map[string]bemfa.Topic{
		config.Bemfa.Switch.Topic: bemfa.NewSwitch(
			func() (err error) {
				err1 := wol.WakeOnLan(config.TargetMac, config.MyIP)
				out, err2 := config.SSH.Command(config.Bemfa.Switch.On)
				if err = errors.Join(err1, err2); err != nil {
					return err
				}

				logger.Debug("switch on", "out", out)
				return nil
			},
			func() (err error) {
				out, err := config.SSH.Command(config.Bemfa.Switch.Off)
				if err != nil {
					return err
				}

				logger.Debug("switch off", "out", out)
				return nil
			},
		),
	}

	b, err := bemfa.New(config.Bemfa.Uid, topics, bemfa.WithLogger(logger.WithGroup("b")))
	if err != nil {
		return nil, fmt.Errorf("bemfa.New: %w", err)
	}

	s := &Server{
		logger: logger,
		bemfa:  b,
	}

	return s, nil
}

func (s *Server) Run() error {
	s.logger.Info("pc start", "listen bemfa", s.bemfa)

	err := s.bemfa.Listen()
	if err != nil {
		return fmt.Errorf("bemfa.Listen: %w", err)
	}

	return nil
}
