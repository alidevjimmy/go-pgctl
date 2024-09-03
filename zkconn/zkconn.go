package zkconn

import (
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
)

// ZKConnection represents a ZooKeeper connection
type ZKConnection struct {
	conn *zk.Conn
}

// Connect establishes a connection to ZooKeeper
func Connect(servers []string, timeout time.Duration) (*ZKConnection, error) {
	conn, _, err := zk.Connect(servers, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ZooKeeper: %w", err)
	}
	return &ZKConnection{conn: conn}, nil
}

// Close closes the ZooKeeper connection
func (zkConn *ZKConnection) Close() {
	if zkConn.conn != nil {
		zkConn.conn.Close()
	}
}

// Create creates a new node in ZooKeeper
func (zkConn *ZKConnection) Create(path string, data []byte) error {
	_, err := zkConn.conn.Create(path, data, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}
	return nil
}

// Get retrieves data from a node in ZooKeeper
func (zkConn *ZKConnection) Get(path string) ([]byte, error) {
	data, _, err := zkConn.conn.Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	return data, nil
}

// Set updates data in a node in ZooKeeper
func (zkConn *ZKConnection) Set(path string, data []byte) error {
	_, err := zkConn.conn.Set(path, data, -1)
	if err != nil {
		return fmt.Errorf("failed to set data: %w", err)
	}
	return nil
}

// Delete removes a node from ZooKeeper
func (zkConn *ZKConnection) Delete(path string) error {
	err := zkConn.conn.Delete(path, -1)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}
	return nil
}

// Exists checks if a node exists in ZooKeeper
func (zkConn *ZKConnection) Exists(path string) (bool, error) {
	exists, _, err := zkConn.conn.Exists(path)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return exists, nil
}
