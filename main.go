package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	fhttp "github.com/Danny-Dasilva/fhttp"
	"github.com/getlantern/systray"
	"golang.org/x/net/webdav"
)

func getAllDrives() ([]string, error) {
	cmd := exec.Command("cmd", "/C", "wmic logicaldisk get name")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var drives []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ":") {
			drives = append(drives, line)
		}
	}
	return drives, nil
}

func main() {
	go systray.Run(SetupTray, onExit)

	err := webDav()
	if err != nil {
		log.Fatalf("WebDAV 服务器启动失败: %v", err)
	}

	proxyService()

	listen()

}

func listen() {
	config := GetConfig()
	address := config.Host + ":" + config.Port
	fmt.Println("监听地址：", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func webDav() error {
	fmt.Println("启动/webdav服务")
	// 获取所有磁盘分区
	drives, err := getAllDrives()
	if err != nil {
		return err
	}
	// 遍历所有盘符，创建对应的 WebDAV 文件系统
	for _, drive := range drives {
		// 创建一个 WebDAV 文件系统处理器
		fmt.Println("创建 WebDAV 文件系统处理器：", drive)
		p := "/webdav/" + drive[0:1] // 路径前缀，如 /C/
		handler := &webdav.Handler{
			Prefix:     p,                        // WebDAV的路径前缀
			FileSystem: webdav.Dir(drive + "//"), // 文件存储的目录
			LockSystem: webdav.NewMemLS(),        // 锁系统，用于防止并发修改
		}
		fmt.Println("webdav服务监听:", p)
		http.Handle(p, handler)
	}
	return nil
}

func proxyService() {
	fmt.Println("启动/proxy监听")

	const ja3 = "771,52393-52392-52244-52243-49195-49199-49196-49200-49171-49172-156-157-47-53-10,65281-0-23-35-13-5-18-16-30032-11-10,29-23-24,0"
	const userAgent = "Chrome Version 57.0.2987.110 (64-bit) Linux"

	http.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		targetURLStr := r.URL.Query().Get("url")
		fmt.Println("有新请求========="+targetURLStr, r.Method)
		// 设置CORS头部
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			// 对于OPTIONS请求，直接返回204状态码和CORS头部
			w.WriteHeader(http.StatusNoContent)
			fmt.Println("预请求处理完成==================")
			return
		}

		if targetURLStr == "" {
			http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
			return
		}

		// 解析目标 URL
		targetURL, err := url.Parse(targetURLStr)
		if err != nil {
			http.Error(w, "Invalid URL parameter", http.StatusBadRequest)
			return
		}

		// 创建请求
		newRequest := &fhttp.Request{
			Method: r.Method,
			URL:    targetURL,
			Header: make(fhttp.Header),
			Body:   r.Body,
		}

		var userAgent_tmp string = userAgent

		for key, values := range r.Header {
			for _, value := range values {
				fmt.Printf("req Header: %s: %s\n", key, value)
				if strings.ToUpper(key) == "USER-AGENT" {
					// userAgent_tmp = value
				} else if key != "Content-Length" {
					newRequest.Header.Add(key, value)
				}
			}
		}

		proxyDialer, err := GetProxyDialer()
		if err != nil {
			http.Error(w, "Failed to create proxy dialer:"+err.Error(), http.StatusInternalServerError)
			return
		}

		client := &fhttp.Client{
			Transport: cycletls.NewTransportWithProxy(ja3, userAgent_tmp, proxyDialer),
		}

		// 发送请求并获取响应
		resp, err := client.Do(newRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("Header: %s: %s\n", key, value)
				w.Header().Add(key, value)
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)

		fmt.Println("处理完成=================================")
	})

}
