package dnssd

import "sync"

type pollable interface {
	init(uintptr) (uintptr, error)
	handleError(error)
}

type pollServerOp struct {
	p   pollable
	ref uintptr
	fd  int
}

var pollServer pollServerState

type pollServerState struct {
	platformPollServerState
	m struct {
		external, internal sync.Mutex
	}
	shared struct {
		ref uintptr
		fd  int
	}
	pollables              map[pollable]*pollServerOp
	pollSlicesUpToDate     bool
	sharedPollableElements []*pollServerOp
	uniquePollableElements []*pollServerOp
}

func (s *pollServerState) startOp(p pollable) error {
	s.m.external.Lock()
	if s.pollables == nil {
		s.pollables = make(map[pollable]*pollServerOp)
	}
	s.stopPoll()
	s.m.internal.Lock()
	defer func() {
		s.m.internal.Unlock()
		s.startPoll()
		s.m.external.Unlock()
	}()
	if _, present := s.pollables[p]; present {
		return ErrStarted
	}
	s.establishSharedConnection()
	ref, err := p.init(s.shared.ref)
	if err != nil {
		return err
	}
	fd := 0
	if s.shared.ref == 0 {
		fd = refSockFd(&ref)
	}
	s.addPollOp(&pollServerOp{p: p, ref: ref, fd: fd})
	return nil
}

func (s *pollServerState) stopOp(p pollable) error {
	s.m.external.Lock()
	defer s.m.external.Unlock()
	s.stopPoll()
	s.m.internal.Lock()
	s.removePollOp(p)
	s.m.internal.Unlock()
	s.startPoll()
	return nil
}

func (s *pollServerState) sharedAndUniquePollables() ([]*pollServerOp, []*pollServerOp) {
	if s.pollSlicesUpToDate {
		return s.sharedPollableElements, s.uniquePollableElements
	}
	s.sharedPollableElements, s.uniquePollableElements = nil, nil
	for _, op := range s.pollables {
		if op.fd > 0 {
			s.uniquePollableElements = append(s.uniquePollableElements, op)
		} else {
			s.sharedPollableElements = append(s.sharedPollableElements, op)
		}
	}
	s.pollSlicesUpToDate = true
	return s.sharedPollableElements, s.uniquePollableElements
}

func (s *pollServerState) removePollOp(p pollable) {
	if op, present := s.pollables[p]; present {
		deallocateRef(&op.ref)
		delete(s.pollables, p)
		s.sharedPollableElements, s.uniquePollableElements = nil, nil
		s.pollSlicesUpToDate = false
	}
}

func (s *pollServerState) addPollOp(p *pollServerOp) {
	s.pollables[p.p] = p
	s.sharedPollableElements, s.uniquePollableElements = nil, nil
	s.pollSlicesUpToDate = false
}

func (s *pollServerState) establishSharedConnection() {
	if len(s.pollables) == 0 && s.shared.ref == 0 {
		if err := createConnection(&s.shared.ref); err != nil {
			_ = err // TODO: do something with err?
			s.shared.ref = 0
		} else {
			s.shared.fd = refSockFd(&s.shared.ref)
			if s.shared.fd < 0 {
				panic("bad fd")
			}
		}
	}
}
