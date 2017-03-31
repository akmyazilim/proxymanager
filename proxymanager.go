package proxymanager

import (
	"fmt"
	"net/url"
	"sync"
)

// DefaultType default proxy type
const DefaultType = "http"

// New return new proxy manager
func New(limit int) *Manager {
	return &Manager{
		List:       make(map[int]Proxy),
		WriteIndex: 0,
		ReadIndex:  0,
		StepIndex:  0,
		Limit:      limit,
	}
}

// Proxy object
type Proxy struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

// Manager  object
type Manager struct {
	sync.Mutex // Embeding olarak ekleyelim.
	List       map[int]Proxy
	WriteIndex int
	ReadIndex  int
	StepIndex  int
	Limit      int
}

// GetWriteIndex return proxy list write index
func (p *Manager) GetWriteIndex() int {
	p.Lock()         // Diğer goroutines'lerin erişmesini engelleyelim.
	defer p.Unlock() // İşlem bittikten sonra erişim engelini kaldıralım
	defer func(p *Manager) {
		p.WriteIndex++
	}(p)
	return p.WriteIndex
}

// Add new Proxy to Proxy List
func (p *Manager) Add(proxy Proxy) {
	if proxy.Type == "" {
		proxy.Type = DefaultType
	}
	p.List[p.GetWriteIndex()] = proxy
}
func (p *Manager) remove(host string) {
	for i, proxy := range p.List {
		if proxy.Host == host {
			delete(p.List, i)
			break
		}
	}
	p.WriteIndex = len(p.List)
}

// Remove proxy
func (p *Manager) Remove(r interface{}) {
	switch r.(type) {
	case string:
		p.remove(r.(string))
	case Proxy:
		p.remove(r.(Proxy).Host)
	}
}

// GiveMeProxy return Proxy from Proxy List
func (p *Manager) GiveMeProxy() Proxy {
	p.Lock()         // Diğer goroutines'lerin erişmesini engelleyelim.
	defer p.Unlock() // İşlem bittikten sonra erişim engelini kaldıralım

	defer func(p *Manager) {
		p.StepIndex++
		if p.StepIndex == p.Limit {
			p.StepIndex = 0
			p.ReadIndex++
		}
	}(p)
	if p.ReadIndex >= p.WriteIndex {
		p.ReadIndex = 0
	}

	return p.List[p.ReadIndex]
}

// GiveMeProxyURL return proxy url
func (p *Manager) GiveMeProxyURL() *url.URL {
	proxy := p.GiveMeProxy()
	userinfo := new(url.Userinfo)
	resultURL := &url.URL{}
	if proxy.Username != "" && proxy.Password != "" {
		userinfo = url.UserPassword(proxy.Username, proxy.Password)
	}

	resultURL.Scheme = proxy.Type
	resultURL.User = userinfo
	resultURL.Host = fmt.Sprintf("%s:%s", proxy.Host, proxy.Port)

	return resultURL
}
