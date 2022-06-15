package group_subscription

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsUnknownMessageType(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "unknown message type",
			args: args{
				err: errors.New(errUnknownMessageType),
			},
			want: true,
		},
		{
			name: "unknown message type false",
			args: args{
				err: errors.New(""),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsUnknownMessageType(tt.args.err), "IsUnknownMessageType(%v)", tt.args.err)
		})
	}
}
