package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alidevjimmy/readyset-replication/query"
	"github.com/alidevjimmy/readyset-replication/util"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Pool struct {
	Nodes  []*Node
	logger *zap.SugaredLogger

	mu sync.Mutex
}

func NewPool(nodes []*Node, logger *zap.SugaredLogger) *Pool {
	return &Pool{
		Nodes:  nodes,
		logger: logger,
		mu:     sync.Mutex{},
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
		if n.Role == Leader {
			return n
		}
	}
	return nil
}

func (p *Pool) SetLeader(idx int) {
	p.Nodes[idx].ChangeRole(Leader)
}

func (p *Pool) GetFollowers() []*Node {
	p.mu.Lock()
	defer p.mu.Unlock()
	var followers []*Node
	for _, n := range p.Nodes {
		if n.Role == Follower {
			followers = append(followers, n)
		}
	}
	return followers
}

func (p *Pool) RunLeaderQueries() {
	l := p.GetLeader()
	if err := p.dropAllSubscriptions(l); err != nil {
		p.logger.Warnf("failed to drop subscriptions: %v", err)
	}
	if err := p.dropReplicationSlots(l); err != nil {
		p.logger.Warnf("failed to drop replication slots: %v", err)
	}
	if err := p.createPublication(l); err != nil {
		p.logger.Warnf("failed to create publication: %v", err)
	}
}

func (p *Pool) RunFollowersQueries() {
	followers := p.GetFollowers()
	leader := p.GetLeader()
	for _, f := range followers {
		if err := p.dropAllSubscriptions(f); err != nil {
			p.logger.Warnf("failed to drop subscriptions: %v", err)
		}
		if err := p.dropReplicationSlots(f); err != nil {
			p.logger.Warnf("failed to drop replication slots: %v", err)
		}
		if err := p.createSubscription(f, leader); err != nil {
			p.logger.Warnf("failed to create subscription: %v", err)
		}
	}
}

func (p *Pool) createPublication(leader *Node) error {
	conn, err := leader.ConnPool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), query.CreatePublicationQuery(leader.ID))
	if err != nil {
		return fmt.Errorf("failed to create publication: %v", err)
	}
	return nil
}

func (p *Pool) dropAllSubscriptions(n *Node) error {
	conn, err := n.ConnPool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %v", err)
	}
	rows, err := conn.Query(context.Background(), "SELECT subname FROM pg_subscription;")
	if err != nil {
		return fmt.Errorf("failed to retrieve subscriptions: %v", err)
	}
	conn.Release()
	defer rows.Close()
	for rows.Next() {
		conn, err := n.ConnPool.Acquire(context.Background())
		if err != nil {
			p.logger.Errorf("failed to acquire connection: %v", err)
		}
		var subName string
		if err := rows.Scan(&subName); err != nil {
			p.logger.Errorf("failed to scan subscription name: %v", err)
		}
		disableSubscriptionSQL := fmt.Sprintf("ALTER SUBSCRIPTION %s DISABLE;", pgx.Identifier{subName}.Sanitize())
		_, err = conn.Exec(context.Background(), disableSubscriptionSQL)
		if err != nil {
			p.logger.Errorf("failed to disable subscription %s: %v", subName, err)
		}
		noneSubscriptionSQL := fmt.Sprintf("ALTER SUBSCRIPTION %s SET (slot_name = NONE)", pgx.Identifier{subName}.Sanitize())
		_, err = conn.Exec(context.Background(), noneSubscriptionSQL)
		if err != nil {
			p.logger.Errorf("failed to set subscription %s slot_name to non: %v", subName, err)
		}
		dropSubscriptionSQL := fmt.Sprintf("DROP SUBSCRIPTION %s", pgx.Identifier{subName}.Sanitize())
		_, err = conn.Exec(context.Background(), dropSubscriptionSQL)
		if err != nil {
			p.logger.Errorf("failed to drop subscription %s: %v", subName, err)
		}
		conn.Release()
	}
	if rows.Err() != nil {
		return fmt.Errorf("error occurred during row iteration: %v", rows.Err())
	}

	return nil
}

func (p *Pool) dropReplicationSlots(n *Node) error {
	conn, err := n.ConnPool.Acquire(context.Background())
	if err != nil {
		p.logger.Errorf("failed to acquire connection: %v", err)
	}
	rows, err := conn.Query(context.Background(), "SELECT slot_name FROM pg_replication_slots WHERE active = false;")
	if err != nil {
		return fmt.Errorf("failed to retrieve slots: %v", err)
	}
	conn.Release()
	defer rows.Close()
	for rows.Next() {
		conn, err := n.ConnPool.Acquire(context.Background())
		if err != nil {
			p.logger.Errorf("failed to acquire connection: %v", err)
		}
		var subName string
		if err := rows.Scan(&subName); err != nil {
			p.logger.Errorf("failed to scan slot name: %v", err)
		}
		dropSlotSQL := fmt.Sprintf("SELECT pg_drop_replication_slot('%s');", subName)
		_, err = conn.Exec(context.Background(), dropSlotSQL)
		if err != nil {
			p.logger.Errorf("failed to drop slot %s: %v", subName, err)
		}
	}
	if rows.Err() != nil {
		return fmt.Errorf("error occurred during row iteration: %v", rows.Err())
	}

	return nil
}

func (p *Pool) createSubscription(follower, leader *Node) error {
	conn, err := follower.ConnPool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()
	subname := fmt.Sprintf("%s_%s_%d", follower.ID, leader.ID, time.Now().Nanosecond())
	_, _, dbname, user, password := util.DSNParse(follower.DSN)
	_, err = conn.Exec(context.Background(), query.CreateSubscriptionQuery(subname, leader.InternalHost, leader.InternalPort, dbname, user, password, leader.ID))
	if err != nil {
		return fmt.Errorf("failed to create subscription: %v", err)
	}
	return nil
}
