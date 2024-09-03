package election

import (
	"math/rand"

	"github.com/alidevjimmy/readyset-replication/node"
)

// Eelect gets elegible nodes and returns the index of the new admin
func Eelect(nodes []*node.Node) int {
	return rand.Intn(len(nodes))
}
