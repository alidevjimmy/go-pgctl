package node

import (
	"sync"

	"github.com/jackc/pgx/v5"
)

type Rule int

const (
	Leader   Rule = 1
	Follower Rule = 2
)

type Node struct {
	Conn *pgx.Conn
	Rule Rule
	ID   string
	DSN  string

	mu sync.Mutex
}

func NewNode(conn *pgx.Conn, id, dsn string) *Node {
	return &Node{
		Conn: conn,
		Rule: Follower,
		DSN:  dsn,
		ID:   id,
		mu:   sync.Mutex{},
	}
}

func (n *Node) ChangeRule(newRule Rule) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Rule = newRule
}
