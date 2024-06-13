package bemfa

import (
	"bytes"
	"errors"
	"fmt"
	"log"
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

// New
// return Bemfa client that subscribe topics
func New(uid string, topics map[string]Topic) (*Bemfa, error) {
	if len(topics) == 0 {
		return nil, errors.New("empty topics")
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("dial error"))
	}

	b := &Bemfa{
		conn:   conn,
		uid:    uid,
		topics: topics,
	}

	var t []string
	for topic := range topics {
		t = append(t, topic)
	}

	err = b.subscribe(t...)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Listen
// Keepalive and listen to msg
func (b *Bemfa) Listen() error {
	go b.Keepalive()

	for {
		err := b.listen()
		if err != nil {
			log.Println(err)
		}
	}
}

// 订阅主题
// cmd=1&uid={{uid}}&topic=xxx1,xxx2,xxx3,xxx4\r\n
//
// 正常返回：
// cmd=1&res=1
func (b *Bemfa) subscribe(topics ...string) error {
	n := len(topics)
	const size = 8 // 单次最多订阅八个主题
	chunks := (n + size - 1) / size
	for chunk := range chunks {
		i := chunk * size
		j := min(i+size, n)

		topic := strings.Join(topics[i:j], ",")

		err := b.write(fmt.Sprintf(`cmd=1&uid=%s&topic=%s`, b.uid, topic))
		if err != nil {
			return errors.Join(err, fmt.Errorf("subscribe topic(%s) error", topic))
		}
	}
	return nil
}

func (b *Bemfa) listen() error {
	buf := make([]byte, 512)
	n, err := b.conn.Read(buf)
	if err != nil {
		return errors.Join(err, fmt.Errorf("read buf error"))
	}

	// 请求过多时，可以考虑使用 channel 以实现读写分离
	return b.handle(buf[:n])
}

func (b *Bemfa) handle(buf []byte) (err error) {
	r := bytes.NewReader(bytes.ReplaceAll(buf, []byte(`&`), []byte(` `)))

	var cmd string
	_, err = fmt.Fscanf(r, "cmd=%s ", &cmd) // ping
	if err != nil {
		return errors.Join(err, fmt.Errorf("read cmd error"))
	}

	switch cmd {
	case cmdPing, cmdSubscribe:
		var res string
		_, err = fmt.Fscanf(r, "res=%s", &res)
		if err != nil {
			return errors.Join(err, fmt.Errorf("scan(&res) error"))
		}

	case cmdPush:
		var uid, topic, msg string
		_, err = fmt.Fscanf(r, "uid=%s topic=%s msg=%s", &uid, &topic, &msg)
		if err != nil {
			return errors.Join(err, fmt.Errorf("scan(&uid, &topic, &msg) error"))
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
			return errors.Join(err, fmt.Errorf("handle msg(%s) error", msg))
		}
	}

	return nil
}

// Keepalive
// Call Ping every 30 seconds
func (b *Bemfa) Keepalive() {
	for {
		err := b.Ping()
		if err != nil {
			// 重试
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
	_, err := b.conn.Write([]byte(req + "\r\n"))
	if err != nil {
		err = errors.Join(err, fmt.Errorf("send req(%s) error", req))
		log.Println("write", err)
		return err
	}

	return nil
}
