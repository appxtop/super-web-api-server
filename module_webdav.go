package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"golang.org/x/net/webdav"
)

func Module_webdav() error {
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
