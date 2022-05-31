package jwt_auth

import (
	"github.com/glide-im/glide/pkg/auth"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestJwtAuthorize_Auth(t *testing.T) {
	type args struct {
		c auth.Info
		t *auth.Token
	}
	tests := []struct {
		name    string
		args    args
		want    *auth.Result
		wantErr bool
	}{
		{
			name:    "nil token",
			args:    args{},
			wantErr: true,
		},
		{
			name: "invalid token",
			args: args{
				c: JwtAuthInfo{
					UID:    "",
					Device: "",
				},
				t: &auth.Token{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := JwtAuthorize{}
			got, err := a.Auth(tt.args.c, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("Auth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Auth() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJwtAuthorize_GetToken(t *testing.T) {
	info := JwtAuthInfo{
		UID:    "3",
		Device: "4",
	}

	impl := NewAuthorizeImpl("secret")

	token, err := impl.GetToken(&info)
	assert.Nil(t, err)
	assert.NotNil(t, token)

	type args struct {
		c auth.Info
	}
	tests := []struct {
		name    string
		args    args
		want    *auth.Token
		wantErr bool
	}{
		{
			name: "nil token",
			args: args{
				c: &info,
			},
			want: token,
		},
		{
			name: "invalid token",
			args: args{
				c: JwtAuthInfo{
					UID:    "",
					Device: "",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := JwtAuthorize{}
			got, err := a.GetToken(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}
