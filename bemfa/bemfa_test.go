package bemfa

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func TestBemfa_Topic(t *testing.T) {
	topic := os.Getenv("topic")
	uid := os.Getenv("uid")

	type args struct {
		uid    string
		topics map[string]Topic
	}

	switchTopic := NewSwitch(
		func() error { log.Printf("here is on \n"); return nil },
		func() error { log.Printf("here is off\n"); return nil },
	)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "switch",
			args: args{
				uid: uid,
				topics: map[string]Topic{
					topic: switchTopic,
				},
			},
			wantErr: false,
		},
	}

	var wg sync.WaitGroup
	for _, tt := range tests {
		wg.Add(1)
		t.Run(tt.name, func(t *testing.T) {
			b, err := New(tt.args.uid, tt.args.topics)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			go func() {
				defer wg.Done()
				err := b.Listen()
				if err != nil {
					panic(err)
				}
			}()
		})
	}
	wg.Wait()
}

func TestBemfa_Listen(t *testing.T) {
	topic := os.Getenv("topic")
	uid := os.Getenv("uid")

	type args struct {
		uid    string
		topics map[string]Topic
	}

	switchTopic := NewSwitch(
		func() error { log.Printf("here is on \n"); return nil },
		func() error { log.Printf("here is off\n"); return nil },
	)

	arg0 := args{
		uid: uid,
		topics: map[string]Topic{
			topic: switchTopic,
		},
	}

	b, err := New(arg0.uid, arg0.topics)
	if err != nil {
		t.Errorf("New() error = %v, wantErr %v", err, nil)
		return
	}

	go b.reconnect()

	go func() {
		for {
			time.Sleep(time.Minute * 2) // make bemfa disconnect
			if err := b.Ping(); err != nil {
				t.Log(err)
			}
		}
	}()

	for {
		if err := b.listen(); err != nil {
			t.Log(err)
		}
	}
}

func TestBemfa_reconnect(t *testing.T) {
	topic := os.Getenv("topic")
	uid := os.Getenv("uid")

	type args struct {
		uid    string
		topics map[string]Topic
	}

	switchTopic := NewSwitch(
		func() error { log.Printf("here is on \n"); return nil },
		func() error { log.Printf("here is off\n"); return nil },
	)

	arg0 := args{
		uid: uid,
		topics: map[string]Topic{
			topic: switchTopic,
		},
	}

	b, err := New(arg0.uid, arg0.topics)
	if err != nil {
		t.Errorf("New() error = %v, wantErr %v", err, nil)
		return
	}

	go b.reconnect()
	time.Sleep(2 * time.Second)

	// 断开连接
	_ = b.conn.Close()

	time.Sleep(2 * time.Second)

	t.Log("begin")
	defer t.Log("finish")

	// 断开连接后，创建并发重连
	wg := sync.WaitGroup{}
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Ping()
		}()
	}

	wg.Wait()
	time.Sleep(5 * time.Second)
}
