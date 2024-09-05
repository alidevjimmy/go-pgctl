package node

import (
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Role int

const (
	Leader   Role = 1
	Follower Role = 2
)

type Node struct {
	ConnPool     *pgxpool.Pool
	Role         Role
	ID           string
	DSN          string
	InternalHost string
	InternalPort string
	
	mu sync.Mutex
}

func NewNode(connPool *pgxpool.Pool, id, dsn, host, port string) *Node {
	return &Node{
		ConnPool:     connPool,
		Role:         Follower,
		DSN:          dsn,
		ID:           id,
		InternalHost: host,
		InternalPort: port,
		mu:           sync.Mutex{},
	}
}

func (n *Node) ChangeRole(newRole Role) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Role = newRole
}
