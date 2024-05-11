package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
)

var fobbidenIp = []string{
	"127.0.0.1",
	"0.0.0.0",
	"117.72.66.215",
}

func main() {
	ip, err := ExternalIP()
	if err != nil {
		fmt.Println(err)
	}
	// tcp 连接，监听 8080 端口
	l, err := net.ListenTCP("tcp4", &net.TCPAddr{Port: 80})
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("start listen ", ip.String(), ":", 8001, "...")

	// 死循环，每当遇到连接时，调用 handle
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handle(client)
	}
}

func handle(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()

	//log.Printf("remote addr: %v\n", client.RemoteAddr())

	// 用来存放客户端数据的缓冲区
	var b [86400]byte
	//从客户端获取数据
	n, err := client.Read(b[:])
	if err != nil {
		log.Println(err)
		return
	}

	var method, URL, address string
	// 从客户端数据读入 method，url
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &URL)
	hostPortURL, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
		return
	}

	// 如果方法是 CONNECT，则为 https 协议
	if method == "CONNECT" {
		address = hostPortURL.Scheme + ":" + hostPortURL.Opaque
	} else { //否则为 http 协议
		address = hostPortURL.Host
		// 如果 host 不带端口，则默认为 80
		if strings.Index(hostPortURL.Host, ":") == -1 { //host 不带端口， 默认 80
			address = hostPortURL.Host + ":80"
		}
	}
	for _, host := range fobbidenIp {
		if hostPortURL.Host == host {
			log.Println("fobbidenIp", address)
			return
		}
	}

	//if strings.Contains(address, "weixin") {
	fmt.Println("n value", n)
	fmt.Println(string(b[:n]))
	fmt.Println("address", address, method)
	//}
	//获得了请求的 host 和 port，向服务端发起 tcp 连接
	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	//如果使用 https 协议，需先向客户端表示连接建立完毕
	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
	} else { //如果使用 http 协议，需将从客户端得到的 http 请求转发给服务端
		server.Write(b[:n])
	}

	//将客户端的请求转发至服务端，将服务端的响应转发给客户端。io.Copy 为阻塞函数，文件描述符不关闭就不停止
	go io.Copy(server, client)
	io.Copy(client, server)
}

// 获取ip
func ExternalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("connected to the network?")
}

// 获取ip
func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}

	return ip
}
