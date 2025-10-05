package httpstub

import (
	"net/http"
	"sync"
)

// Storage is an in-memory storage for HTTP stubs.
type Storage struct {
	// represents [URL][Method]
	stubs map[string]map[string]Response

	m sync.Mutex
}

// NewStorage creates a new instance of Storage.
func NewStorage() *Storage {
	return &Storage{
		stubs: map[string]map[string]Response{},
		m:     sync.Mutex{},
	}
}

// Add adds a new ProtoStub to the storage.
func (p *Storage) Add(s Stub) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.stubs[s.Path] == nil {
		p.stubs[s.Path] = map[string]Response{}
	}
	p.stubs[s.Path][s.Method] = s.Response
}

// Get retrieves the Output for a given URL and method.
func (p *Storage) Get(req *http.Request) (Response, error) {
	p.m.Lock()
	defer p.m.Unlock()

	matchingURLs, ok := p.stubs[req.URL.Path]
	if !ok {
		return Response{}, ErrStubNotFound
	}

	stub, ok := matchingURLs[req.Method]
	if !ok {
		return Response{}, ErrMethodNotAllowed
	}

	return stub, nil
}
