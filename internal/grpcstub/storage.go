package grpcstub

import (
	"encoding/json"
	"sync"
)

type protoStorage struct {
	// represents [serviceName][methodName]
	stubs map[string]map[string]Output

	m sync.Mutex
}

var _ Repository = &protoStorage{}

func NewStorage() *protoStorage {
	return &protoStorage{
		stubs: map[string]map[string]Output{},
		m:     sync.Mutex{},
	}
}

func (p *protoStorage) Add(s ProtoStub) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.stubs[s.Service] == nil {
		p.stubs[s.Service] = map[string]Output{}
	}
	p.stubs[s.Service][s.Method] = s.Output
}

func (p *protoStorage) Get(service string, method string, in json.RawMessage) (Output, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	s, ok := p.stubs[service][method]
	if !ok {
		return Output{}, false
	}

	return s, true
}
