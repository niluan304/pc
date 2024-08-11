package bemfa

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"
)

// Bemfa
// api doc: https://cloud.bemfa.com/docs/src/tcp.html
type Bemfa struct {
	conn   net.Conn
	uid    string
	topics map[string]Topic

	disconnect chan struct{}

	logger *slog.Logger
}

const addr = "bemfa.com:8344"

const (
	cmdPing              = "0" // ping 的响应
	cmdSubscribe         = "1" // 订阅消息，当设备发送一次此消息类型，之后就可以收到发往该主题的消息
	cmdPush              = "2" // 发布消息，向订阅该主题的设备发送消息
	cmdSubscribeWithLast = "3" // 订阅消息，和cmd=1相同，并且会拉取一次已发送过的消息
	cmdTime              = "7" // 获取时间，获取当前北京时间
	cmdLast              = "9" // 遗嘱消息，拉取一次已经发送的消息
)

type Option func(*Bemfa)

func WithLogger(logger *slog.Logger) Option {
	return func(b *Bemfa) {
		if logger == nil {
			return
		}

		b.logger = logger
	}
}

// New
// return Bemfa client that subscribe topics
func New(uid string, topics map[string]Topic, opts ...Option) (*Bemfa, error) {
	if len(topics) == 0 {
		return nil, errors.New("empty topics")
	}

	b := &Bemfa{
		uid:        uid,
		topics:     topics,
		disconnect: make(chan struct{}),
		logger:     slog.Default(),
	}

	for _, opt := range opts {
		opt(b)
	}

	return b, nil
}

// Listen
// keepalive and listen to msg
func (b *Bemfa) Listen() error {
	if err := b.subscribe(); err != nil {
		return err
	}

	go b.reconnect()
	go b.keepalive()

	for {
		err := b.listen()
		if err != nil {
			b.logger.Error("Listen", "err", err)
		}
	}
}

// 触发 disconnect 信号时，重新连接 bemfa.com
func (b *Bemfa) reconnect() {
	last := time.Now()

	for {
		<-b.disconnect

		// 2s 内最多重连一次
		now := time.Now()
		if last.Add(2 * time.Second).After(now) {
			continue
		}
		last = now

		// 匿名函数里执行任务，避免写 if err != nil { return }, 跳出 for 循环
		func() {
			b.logger.Warn("reconnect when disconnect")

			if err := b.subscribe(); err != nil {
				b.logger.Error("reconnect with subscribe", "err", err)
				return
			}
		}()
	}
}

// 订阅主题
// cmd=1&uid={{uid}}&topic=xxx1,xxx2,xxx3,xxx4\r\n
//
// 正常返回：
// cmd=1&res=1
func (b *Bemfa) subscribe() error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("fail to dial %s, err: %w", addr, err)
	}

	b.conn = conn

	var topics []string
	for topic := range b.topics {
		topics = append(topics, topic)
	}

	n := len(topics)
	const size = 8 // 单次最多订阅八个主题
	chunks := (n + size - 1) / size
	for chunk := range chunks {
		i := chunk * size
		j := min(i+size, n)

		topic := strings.Join(topics[i:j], ",")

		err := b.write(fmt.Sprintf(`cmd=%s&uid=%s&topic=%s`, cmdSubscribe, b.uid, topic))
		if err != nil {
			return fmt.Errorf("fail to subscribe topic(%s), err: %w", topic, err)
		}
	}
	return nil
}

func (b *Bemfa) listen() error {
	buf := make([]byte, 512)
	n, err := b.conn.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		if t, ok := err.(temporary); ok && t.Temporary() {
			time.Sleep(time.Second)
			return nil
		}

		b.disconnect <- struct{}{}
		return fmt.Errorf("conn fail to read buf, err: %w", err)
	}

	// 请求过多时，可以考虑使用 channel 以实现读写分离
	return b.handle(buf[:n])
}

func (b *Bemfa) handle(buf []byte) (err error) {
	b.logger.Debug("handle", "msg", string(buf))

	r := bytes.NewReader(bytes.ReplaceAll(buf, []byte(`&`), []byte(` `)))

	var cmd string
	_, err = fmt.Fscanf(r, "cmd=%s ", &cmd) // ping
	if err != nil {
		return fmt.Errorf("handle fail to read cmd from: %s, err: %w", buf, err)
	}

	switch cmd {
	case cmdPing, cmdSubscribe:
		var res string
		_, err = fmt.Fscanf(r, "res=%s", &res)
		if err != nil {
			return fmt.Errorf("fail to scan(&res) from: %s, err: %w", buf, err)
		}

	case cmdPush:
		var uid, topic, msg string
		_, err = fmt.Fscanf(r, "uid=%s topic=%s msg=%s", &uid, &topic, &msg)
		if err != nil {
			return fmt.Errorf("fail to scan(&uid, &topic, &msg) from: %s, err: %w", buf, err)
		}

		if uid != b.uid {
			return fmt.Errorf("uid don't match, got: %s", uid)
		}

		t := b.topics[topic]
		if t == nil {
			return fmt.Errorf("topic(%s) not found", topic)
		}

		err = t.Handle(msg)
		if err != nil {
			return fmt.Errorf("fail to handle msg(%s), err: %w", msg, err)
		}
	}

	return nil
}

// keepalive
// Call Ping every 30 seconds
func (b *Bemfa) keepalive() {
	for {
		if err := b.Ping(); err != nil {
			b.logger.Error("keepalive", "err", err)
			time.Sleep(time.Second * 2)
			continue
		}
		time.Sleep(time.Second * 30)
	}
}

// Ping
// send request `ping` to bemfa, and the response will be `cmd=0&res=1`
func (b *Bemfa) Ping() error {
	err := b.write(`ping`)
	if err != nil {
		return err
	}
	return nil
}

// cmd=1 时为订阅消息，当设备发送一次此消息类型，之后就可以收到发往该主题的消息
//
// cmd=2 时为发布消息，向订阅该主题的设备发送消息
//
// cmd=3 是订阅消息，和cmd=1相同，并且会拉取一次已发送过的消息
//
// cmd=7 是获取时间，获取当前北京时间
//
// cmd=9 为遗嘱消息，拉取一次已经发送的消息
func (b *Bemfa) write(req string) error {
write:
	b.logger.Debug("write", "req", req)

	_, err := b.conn.Write([]byte(req + "\r\n"))
	if err != nil {
		if t, ok := err.(temporary); ok && t.Temporary() {
			time.Sleep(time.Second)
			goto write
		}

		_ = b.conn.Close()
		b.disconnect <- struct{}{}
		return fmt.Errorf("fail to write(%s), err: %w", req, err)
	}

	return nil
}

type temporary interface {
	Temporary() bool
}
