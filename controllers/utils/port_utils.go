package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

type PortData struct {
	Namespace string
	Name      string
	Port      int32
}

type PortMap struct {
	proxyPorts map[int32]PortData
	services   map[string]int32
}

func (*PortMap) GetSvcKey(data PortData) string {
	return fmt.Sprintf("%s:%s:%d", data.Name, data.Namespace, data.Port)
}

func (p *PortMap) AddPort(port int32, data PortData) {
	p.proxyPorts[port] = data
	p.services[p.GetSvcKey(data)] = port
}

func (p *PortMap) RemovePort(port int32) {
	if _, ok := p.proxyPorts[port]; ok {
		delete(p.services, p.GetSvcKey(p.proxyPorts[port]))
		delete(p.proxyPorts, port)
	}
}

func (p *PortMap) RemoveSvc(data PortData) {
	if port, ok := p.services[p.GetSvcKey(data)]; ok {
		delete(p.proxyPorts, port)
		delete(p.services, p.GetSvcKey(data))
	}
}

func (p *PortMap) GetPort(data PortData) *int32 {
	if port, ok := p.services[p.GetSvcKey(data)]; ok {
		return &port
	}

	return nil
}

func (p *PortMap) GetSvc(port int32) *PortData {
	if svc, ok := p.proxyPorts[port]; ok {
		return &svc
	}

	return nil
}

func (p *PortMap) portExist(port int32) bool {
	_, ok := p.proxyPorts[port]
	return ok
}

func (p *PortMap) SvcExist(data PortData) bool {
	_, ok := p.services[p.GetSvcKey(data)]
	return ok
}

func (p *PortMap) GetRandomPort(o PortMap) int32 {
	var port int32
	for {
		port = 1024 + rand.Int31n(65535-1024)
		if !p.portExist(port) && !o.portExist(port) {
			return port
		}
	}
}

func (p *PortMap) ParseBytes(data []byte) error {
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	return nil
}

func (p *PortMap) ToBytes() ([]byte, error) {
	return json.Marshal(p)
}

func (p *PortMap) IsEquals(o PortMap) bool {
	if len(p.proxyPorts) != len(o.proxyPorts) || len(p.services) != len(o.services) {
		return false
	}

	for k, v := range p.proxyPorts {
		if o.proxyPorts[k] != v {
			return false
		}
	}

	for k, v := range p.services {
		if o.services[k] != v {
			return false
		}
	}

	return true
}
