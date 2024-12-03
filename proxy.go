package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows/registry"

	"golang.org/x/net/proxy"
)

// 从注册表获取代理信息
func GetProxy() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.READ)
	if err != nil {
		return "", fmt.Errorf("打开注册表失败: %w", err)
	}
	defer k.Close()

	ProxyEnable, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil {
		return "", fmt.Errorf("获取ProxyEnable失败: %w", err)
	}
	fmt.Println("ProxyEnable:", ProxyEnable)
	if ProxyEnable != 1 {
		return "", nil
	}

	ProxyServer, _, err := k.GetStringValue("ProxyServer")
	if err != nil {
		return "", fmt.Errorf("获取ProxyServer失败: %w", err)
	}

	return ProxyServer, nil
}

// 获取代理,如果没有获取到就返回proxy.Direct
func GetProxyDialer() (proxy.ContextDialer, error) {
	ProxyServer, err := GetProxy()
	if err != nil || ProxyServer == "" {
		log.Printf("没有获取到代理,使用直连: %v", err)
		return proxy.Direct, nil
	}
	fmt.Println("使用代理:--->" + ProxyServer)
	proxyDialer, err := NewConnectDialer("http://"+ProxyServer, "ua by super-web-api")
	return proxyDialer, err
}
