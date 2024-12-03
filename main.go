package main

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	http "github.com/Danny-Dasilva/fhttp"
	"github.com/getlantern/systray"
)

const ja3 = "771,52393-52392-52244-52243-49195-49199-49196-49200-49171-49172-156-157-47-53-10,65281-0-23-35-13-5-18-16-30032-11-10,29-23-24,0"
const userAgent = "Chrome Version 57.0.2987.110 (64-bit) Linux"

// 复制头部

func main() {

	go systray.Run(SetupTray, onExit)

	fmt.Println("程序开始运行...")
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
		newRequest := &http.Request{
			Method: r.Method,
			URL:    targetURL,
			Header: make(http.Header),
			Body:   r.Body,
		}

		var userAgent_tmp string = userAgent

		for key, values := range r.Header {
			for _, value := range values {
				fmt.Printf("req Header: %s: %s\n", key, value)
				if strings.ToUpper(key) == "USER-AGENT" {
					userAgent_tmp = value
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

		client := &http.Client{
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

	config := GetConfig()

	address := config.Host + ":" + config.Port

	fmt.Println("监听地址：", address)

	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}
