package utils

import "net"

/*
* @Author: zouyx
* @Email: 1003941268@qq.com
* @Date:   2025/11/20 上午11:18
* @Package:
 */

// GetLocalIP 获取本机的局域网IPv4地址（优先返回非回环、非虚拟网卡）
// 如果获取失败，返回 "localhost"
func GetLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}
	for _, iface := range interfaces {
		// 排除未启用、loopback 网卡
		if iface.Flags&(net.FlagUp|net.FlagLoopback) != net.FlagUp {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// 过滤 IPv6 / 回环 / 非局域网地址
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			// 返回 IPv4 地址
			return ip.String()
		}
	}
	return "localhost"
}
