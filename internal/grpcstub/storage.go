package grpcstub

import (
	"encoding/json"
	"sync"
)

type ProtoStorage struct {
	// represents [serviceName][methodName]
	stubs map[string]map[string]Output

	m sync.Mutex
}

var _ Repository = &ProtoStorage{}

func NewStorage() *ProtoStorage {
	return &ProtoStorage{
		stubs: map[string]map[string]Output{},
		m:     sync.Mutex{},
	}
}

func (p *ProtoStorage) Add(s ProtoStub) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.stubs[s.Service] == nil {
		p.stubs[s.Service] = map[string]Output{}
	}
	p.stubs[s.Service][s.Method] = s.Output
}

func (p *ProtoStorage) Get(service string, method string, _ json.RawMessage) (Output, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	s, ok := p.stubs[service][method]
	if !ok {
		return Output{}, false
	}

	return s, true
}
