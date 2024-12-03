package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func main() {
	switch runtime.GOOS {
	case "windows":
		getWindowsProxy()
	case "darwin":
		getMacProxy()
	case "linux":
		getLinuxProxy()
	default:
		log.Fatalf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// 获取 Windows 上的代理设置
func getWindowsProxy() {
	// 打开注册表项
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatalf("无法打开注册表项: %v", err)
	}
	defer key.Close()

	// 获取代理启用状态
	proxyEnable, _, err := key.GetIntegerValue("ProxyEnable")
	if err != nil {
		log.Fatalf("无法获取 ProxyEnable 键值: %v", err)
	}

	// 如果代理启用，则获取代理服务器地址
	if proxyEnable == 1 {
		proxyServer, _, err := key.GetStringValue("ProxyServer")
		if err != nil {
			log.Fatalf("无法获取 ProxyServer 键值: %v", err)
		}
		fmt.Printf("Windows代理服务器地址: %s\n", proxyServer)
	} else {
		fmt.Println("Windows上没有启用代理")
	}
}

// 获取 macOS 上的代理设置
func getMacProxy() {
	// 执行 networksetup 命令来获取代理设置
	cmd := exec.Command("networksetup", "-getwebproxy", "Wi-Fi") // 使用 Wi-Fi 接口，或者替换为 Ethernet
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("无法获取代理设置: %v", err)
	}

	// 打印命令输出
	outputStr := string(output)
	fmt.Println("macOS代理设置输出:", outputStr)

	// 检查是否启用代理
	if strings.Contains(outputStr, "Enabled: Yes") {
		// 获取代理服务器地址和端口
		proxyInfo := strings.Split(outputStr, "\n")
		for _, line := range proxyInfo {
			if strings.Contains(line, "Server") {
				fmt.Println("macOS代理服务器地址:", strings.TrimSpace(line))
			} else if strings.Contains(line, "Port") {
				fmt.Println("macOS代理端口:", strings.TrimSpace(line))
			}
		}
	} else {
		fmt.Println("macOS上没有启用代理")
	}
}

// 获取 Linux 上的代理设置
func getLinuxProxy() {
	// 检查环境变量中的代理设置
	httpProxy := os.Getenv("http_proxy")
	httpsProxy := os.Getenv("https_proxy")

	if httpProxy != "" || httpsProxy != "" {
		fmt.Println("检测到Linux代理设置:")
		if httpProxy != "" {
			fmt.Println("HTTP代理:", httpProxy)
		}
		if httpsProxy != "" {
			fmt.Println("HTTPS代理:", httpsProxy)
		}
	} else {
		// 如果没有环境变量设置，检查 GNOME 桌面环境的代理设置
		cmd := exec.Command("gsettings", "get", "org.gnome.system.proxy", "mode")
		output, err := cmd.CombinedOutput()
		if err == nil && strings.TrimSpace(string(output)) == "'manual'" {
			cmd = exec.Command("gsettings", "get", "org.gnome.system.proxy", "http-host")
			proxyOutput, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalf("无法获取 GNOME HTTP 代理: %v", err)
			}
			proxyAddress := strings.TrimSpace(string(proxyOutput))
			fmt.Printf("GNOME代理服务器地址: %s\n", proxyAddress)
		} else {
			fmt.Println("Linux上没有启用代理")
		}
	}
}
