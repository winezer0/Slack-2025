package portscan

import (
	"context"
	"fmt"
	"net"
	"slack-wails/core/webscan"
	"slack-wails/lib/clients"
	"slack-wails/lib/gologger"
	"slack-wails/lib/gonmap"
	"slack-wails/lib/structs"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func TcpScan(ctx, ctrlCtx context.Context, taskId string, addresses <-chan Address, workers, timeout int, proxy clients.Proxy) {
	var id int32
	single := make(chan struct{})
	retChan := make(chan *structs.InfoResult)
	var wg sync.WaitGroup
	openPorts := make(map[string]bool) // 记录开放的端口
	go func() {
		for pr := range retChan {
			runtime.EventsEmit(ctx, "webFingerScan", pr)
		}
		close(single)
	}()
	// port scan func
	portScan := func(add Address) {
		defer wg.Done()
		defer func() {
			atomic.AddInt32(&id, 1)
			runtime.EventsEmit(ctx, "progressID", id)
		}()
		if ctrlCtx.Err() != nil {
			return
		}
		pr := Connect(ctx, taskId, add.IP, add.Port, timeout, proxy)
		// atomic.AddInt32(&id, 1)
		// runtime.EventsEmit(ctx, "progressID", id)
		if pr == nil {
			return
		}
		// 检查1-10端口的开放情况
		if pr.Port >= 1 && pr.Port <= 20 {
			openPorts[pr.Host] = true // 记录该IP有开放端口
			gologger.IntervalError(ctx, fmt.Sprintf("[portscan] %s 疑似cdn地址，会对未识别到服务的端口进行过滤", pr.Host))
		} else if openPorts[pr.Host] && pr.Scheme == "" {
			// 如果该IP在1-10端口有开放，后续端口必须识别到服务
			return // 如果没有识别到服务，则不返回
		}
		retChan <- pr
	}
	threadPool, _ := ants.NewPoolWithFunc(workers, func(ipaddr interface{}) {
		ipa := ipaddr.(Address)
		portScan(ipa)
	})
	defer threadPool.Release()
	for add := range addresses {
		if ctrlCtx.Err() != nil {
			return
		}
		wg.Add(1)
		threadPool.Invoke(add)
	}
	wg.Wait()
	close(retChan)
	<-single
}

type Address struct {
	IP   string
	Port int
}

func Connect(ctx context.Context, taskId, ip string, port, timeout int, proxy clients.Proxy) *structs.InfoResult {
	scanner := gonmap.New()
	status, response := scanner.Scan(ip, port, time.Second*time.Duration(timeout), proxy)

	// 端口关闭或未知，直接返回 nil
	if status == gonmap.Closed || status == gonmap.Unknown {
		return nil
	}
	var tcpfinger []string
	// 默认协议设为 unknow
	scheme := "unknow"
	if response != nil && response.FingerPrint.Service != "" {
		scheme = response.FingerPrint.Service
	}
	tcpinfo := &webscan.WebInfo{
		Protocol: scheme,
		Banner:   strings.ToLower(response.Raw),
	}
	if scheme == "http" || scheme == "https" {
		tcpfinger = webscan.Scan(ctx, tcpinfo, webscan.FingerprintDB)
	}
	result := &structs.InfoResult{
		TaskId:       taskId,
		Host:         ip,
		Port:         port,
		Scheme:       scheme,
		URL:          fmt.Sprintf("%s://%s:%d", scheme, ip, port),
		Fingerprints: tcpfinger,
		Detect:       "Default",
	}

	// 若是 HTTP/HTTPS，尝试请求获取状态码
	if scheme == "http" || scheme == "https" {
		resp, err := clients.SimpleGet(result.URL, clients.NewRestyClient(nil, true))
		if err == nil {
			result.StatusCode = resp.StatusCode()
		}
	}

	return result
}

func WrapperTcpWithTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := &net.Dialer{Timeout: timeout}
	return WrapperTCP(network, address, d)
}

func WrapperTCP(network, address string, forward *net.Dialer) (net.Conn, error) {
	//get conn
	conn, err := forward.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
