package main

import (
	"testing"

	"github.com/posteo/go-agentx/value"
)

func Test_compareOids(t *testing.T) {
	type args struct {
		x value.OID
		y value.OID
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "test equal", args: args{value.OID{1, 1, 1}, value.OID{1, 1, 1}}, want: 0},
		{name: "eq size first less", args: args{value.OID{1, 1, 0}, value.OID{1, 1, 1}}, want: -1},
		{name: "eq size first more", args: args{value.OID{1, 2, 1}, value.OID{1, 1, 1}}, want: 1},

		{name: "neq size first less", args: args{value.OID{1, 1, 0}, value.OID{1, 1, 10, 10}}, want: -1},
		{name: "neq size first more", args: args{value.OID{1, 3}, value.OID{1, 1, 1, 1}}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareOids(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("compareOids() = %v, want %v", got, tt.want)
			}
		})
	}
}
