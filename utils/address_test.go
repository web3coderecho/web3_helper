package utils

import (
	"strings"
	"testing"
)

func TestTronToEth(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				address: "TJ7hhYhVhaxNx6BPyq7yFpqZrQULL3JSdb",
			},
			want: "0x595C4A379AB80C202F0372BBF9BBF3FAD6CA8768",
		},
		{
			name: "test2",
			args: args{
				address: "TX94b5x9C16JUFVgN7SQZcYHX4uWEdcV48",
			},
			want: "0xE837C72A310F201AD2E3CA4E44C1DD0D41F4EC8B",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TronToEth(tt.args.address); !strings.EqualFold(got, tt.want) {
				t.Errorf("TronToEth() = %v, want %v", got, tt.want)
			}
		})
	}
}
