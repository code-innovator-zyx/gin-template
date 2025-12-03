package jwt

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/20 上午9:28
* @Package:
 */

func TestHash(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test hash",
			args: args{token: "test hash"},
			want: "54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Hash(tt.args.token), "Hash(%v)", tt.args.token)
		})
	}
}

func TestSecureCompare(t *testing.T) {
	type args struct {
		a string
		b string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test compare",
			args: args{
				a: "54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b",
				b: "54a6483b8aca55c9df2a35baf71d9965ddfd623468d81d51229bd5eb7d1e1c1b",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, SecureCompare(tt.args.a, tt.args.b), "SecureCompare(%v, %v)", tt.args.a, tt.args.b)
		})
	}
}
