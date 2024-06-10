package wol

import (
	"testing"
)

func TestWakeOnLan(t *testing.T) {
	type args struct {
		mac string
		ip  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "power on",
			args: args{
				mac: "9c:6b:00:5f:a7:35",
				ip:  "192.168.1.219", // 主机自己的IP
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WakeOnLan(tt.args.mac, tt.args.ip); (err != nil) != tt.wantErr {
				t.Errorf("WakeOnLan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
