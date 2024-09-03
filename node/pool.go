package node

import (
	"sync"
)

type Pool struct {
	Nodes []*Node
	mu    sync.Mutex
}

func NewPool(nodes []*Node) *Pool {
	return &Pool{
		Nodes: nodes,
		mu:    sync.Mutex{},
	}
}

func (p *Pool) AddNode(n *Node) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Nodes = append(p.Nodes, n)
}

func (p *Pool) RemoveNode(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, n := range p.Nodes {
		if n.ID == id {
			p.Nodes = append(p.Nodes[:i], p.Nodes[i+1:]...)
			return
		}
	}
}

func (p *Pool) GetNode(id string) *Node {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, n := range p.Nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
}

func (p *Pool) GetLeader() *Node {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, n := range p.Nodes {
		if n.Rule == Leader {
			return n
		}
	}
	return nil
}

func (p *Pool) SetLeader(idx int) {
	p.Nodes[idx].ChangeRule(Leader)
}
