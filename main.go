package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

const setPath = "/liuyongproxy"
const username = "liuyong"
const passwrod = "12345678"

var fobbidenIp = []string{
	"127.0.0.1",
	"0.0.0.0",
	"117.72.66.215",
}
var allowIp []string

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
	fmt.Println("start listen ", ip.String(), ":", 80, "...")

	// 死循环，每当遇到连接时，调用 handle
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handle(client)
	}
}
func setWhiteIp() (err error) {
	ip, err := ExternalIP()
	if err != nil {
		return
	}
	if len(allowIp) > 5 {
		allowIp = []string{}
	}
	allowIp = append(allowIp, ip.String())
	fmt.Println("set ip success", allowIp, ip.String())
	return
}
func checkWhiteIp() (pass bool, ipvalue string, err error) {

	ip, err := ExternalIP()
	if err != nil {
		return
	}
	ipvalue = ip.String()
	for _, v := range allowIp {
		if v == ipvalue {
			pass = true
			return
		}
	}
	return
}

func handle(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()
	reader := bufio.NewReader(client)
	request, err := http.ReadRequest(reader)
	if err != nil {
		return
	}

	// 创建新的请求对象
	newRequest := &http.Request{
		Method: request.Method,
		URL:    request.URL,
		Header: make(http.Header),
	}
	if request.URL.Path == setPath {
		values := request.URL.Query()
		if values.Get("username") == username && values.Get("password") == passwrod {
			err = setWhiteIp()
			if err != nil {
				fmt.Println("setWhiteIp err:", err)
			}
		}
		return
	}
	//检测ip白名单
	pass, ip, err := checkWhiteIp()
	if err != nil {
		fmt.Println("checkWhiteIp err:", err)
		return
	}
	if !pass {
		fmt.Println("checkWhiteIp fail,ip:", ip)
		return
	}
	fmt.Println("white list ", allowIp)
	fmt.Println("checkWhiteIp pass,ip:", ip)

	// 确保URL是绝对的，这里简单示例未做转换，实际可能需要更复杂的处理逻辑
	if newRequest.URL.Scheme == "" || newRequest.URL.Host == "" {
		fmt.Println("scheme,host 不能为空", err)
		return
	}

	// 复制头部信息，注意过滤掉特定头部
	for key, values := range request.Header {
		if key != "Authorization" && key != "Transfer-Encoding" && key != "Content-Length" {
			for _, value := range values {
				newRequest.Header.Add(key, value)
			}
		}
	}

	// 发送请求并处理响应
	resp, err := http.DefaultTransport.RoundTrip(newRequest)
	if err != nil {
		log.Println("转发请求时出错:", err)
		return
	}
	defer resp.Body.Close()

	// 写入响应状态行
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	if _, err := client.Write([]byte(statusLine)); err != nil {
		log.Println("写入状态行时出错:", err)
		return
	}

	// 写入响应头
	for key, values := range resp.Header {
		for _, value := range values {
			headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
			if _, err := client.Write([]byte(headerLine)); err != nil {
				log.Println("写入响应头时出错:", err)
				return
			}
		}
	}

	// 写入空行和响应体
	if _, err := client.Write([]byte("\r\n")); err != nil {
		log.Println("写入空行时出错:", err)
		return
	}
	if _, err := io.Copy(client, resp.Body); err != nil {
		log.Println("复制响应体时出错:", err)
	}

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
