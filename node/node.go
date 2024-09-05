package node

import (
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Rule int

const (
	Leader   Rule = 1
	Follower Rule = 2
)

type Node struct {
	ConnPool     *pgxpool.Pool
	Rule         Rule
	ID           string
	DSN          string
	InternalHost string
	InternalPort string

	mu sync.Mutex
}

func NewNode(connPool *pgxpool.Pool, id, dsn, host, port string) *Node {
	return &Node{
		ConnPool: connPool,
		Rule:     Follower,
		DSN:      dsn,
		ID:       id,
		InternalHost: host,
		InternalPort: port,
		mu:       sync.Mutex{},
	}
}

func (n *Node) ChangeRule(newRule Rule) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Rule = newRule
}
