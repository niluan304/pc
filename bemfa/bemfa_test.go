package bemfa_test

import (
	"log"
	"os"
	"sync"
	"testing"

	"github.com/nilluan304/pc/bemfa"
)

func TestBemfa_Topic(t *testing.T) {
	topic := os.Getenv("topic")
	uid := os.Getenv("uid")

	type args struct {
		uid    string
		topics map[string]bemfa.Topic
	}

	switchTopic := bemfa.NewSwitch(
		func() error {
			log.Printf("here is on\n")
			return nil
		},
		func() error {
			log.Printf("here is off\n")
			return nil
		},
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
				topics: map[string]bemfa.Topic{
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
			b, err := bemfa.New(tt.args.uid, tt.args.topics)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			go func() {
				b.Listen()
				wg.Done()
			}()
		})
	}
	wg.Wait()
}
