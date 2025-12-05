package utils

import "testing"

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025 2025/12/5 下午12:03
* @Package:
 */

func TestGetLocalIP(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{"test", GetLocalIP()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLocalIP(); got != tt.want {
				t.Errorf("GetLocalIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
