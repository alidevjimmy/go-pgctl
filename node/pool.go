package node

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alidevjimmy/readyset-replication/query"
	"github.com/alidevjimmy/readyset-replication/util"
	"github.com/jackc/pgx/v5"
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
	p.dropAllSubscriptions(l)
	p.dropReplicationSlots(l)
	p.createPublication(l)
}

func (p *Pool) RunFollowersQueries() {
	followers := p.GetFollowers()
	leader := p.GetLeader()
	for _, f := range followers {
		p.dropAllSubscriptions(f)
		p.dropReplicationSlots(f)
		p.createSubscription(f, leader)
	}
}

func (p *Pool) createPublication(leader *Node) {
	conn, err := leader.ConnPool.Acquire(context.Background())
	if err != nil {
		log.Printf("[ERROR] failed to acquire connection: %v", err)
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), query.CreatePublicationQuery(leader.ID))
	if err != nil {
		log.Printf("[WARNING] failed to create publication: %v", err)
	}
	log.Printf("[INFO] Publication created for node %s", leader.ID)
}

func (p *Pool) dropAllSubscriptions(n *Node) {
	conn, err := n.ConnPool.Acquire(context.Background())
	if err != nil {
		log.Printf("[ERROR] failed to acquire connection: %v", err)
	}
	rows, err := conn.Query(context.Background(), "SELECT subname FROM pg_subscription;")
	if err != nil {
		log.Printf("Failed to retrieve subscriptions: %v\n", err)
	}
	conn.Release()
	defer rows.Close()
	for rows.Next() {
		conn, err := n.ConnPool.Acquire(context.Background())
		if err != nil {
			log.Printf("[ERROR] failed to acquire connection: %v", err)
		}
		var subName string
		if err := rows.Scan(&subName); err != nil {
			log.Printf("Failed to scan subscription name: %v\n", err)
		}
		disableSubscriptionSQL := fmt.Sprintf("ALTER SUBSCRIPTION %s DISABLE;", pgx.Identifier{subName}.Sanitize())
		_, err = conn.Exec(context.Background(), disableSubscriptionSQL)
		if err != nil {
			log.Printf("Failed to disable subscription %s: %v\n", subName, err)
		}
		noneSubscriptionSQL := fmt.Sprintf("ALTER SUBSCRIPTION %s SET (slot_name = NONE)", pgx.Identifier{subName}.Sanitize())
		_, err = conn.Exec(context.Background(), noneSubscriptionSQL)
		if err != nil {
			log.Printf("Failed to set subscription %s slot_name to non: %v\n", subName, err)
		}
		dropSubscriptionSQL := fmt.Sprintf("DROP SUBSCRIPTION %s", pgx.Identifier{subName}.Sanitize())
		_, err = conn.Exec(context.Background(), dropSubscriptionSQL)
		if err != nil {
			log.Printf("Failed to drop subscription %s: %v\n", subName, err)
		}

		conn.Release()
		fmt.Printf("Dropped subscription: %s\n", subName)
	}
	if rows.Err() != nil {
		log.Printf("Error occurred during row iteration: %v\n", rows.Err())
	}
}

func (p *Pool) dropReplicationSlots(n *Node) {
	conn, err := n.ConnPool.Acquire(context.Background())
	if err != nil {
		log.Printf("[ERROR] failed to acquire connection: %v", err)
	}
	rows, err := conn.Query(context.Background(), "SELECT slot_name FROM pg_replication_slots WHERE active = false;")
	if err != nil {
		log.Printf("Failed to retrieve slots: %v\n", err)
	}
	conn.Release()
	defer rows.Close()
	for rows.Next() {
		conn, err := n.ConnPool.Acquire(context.Background())
		if err != nil {
			log.Printf("[ERROR] failed to acquire connection: %v", err)
		}
		var subName string
		if err := rows.Scan(&subName); err != nil {
			log.Printf("Failed to scan slot name: %v\n", err)
		}
		dropSlotSQL := fmt.Sprintf("SELECT pg_drop_replication_slot('%s');", subName)
		_, err = conn.Exec(context.Background(), dropSlotSQL)
		if err != nil {
			log.Printf("Failed to drop slot %s: %v\n", subName, err)
		}
		fmt.Printf("Dropped slot: %s\n", subName)
	}
	if rows.Err() != nil {
		log.Printf("Error occurred during row iteration: %v\n", rows.Err())
	}

	log.Printf("All slots for %s dropped successfully!", n.ID)
}

func (p *Pool) createSubscription(follower, leader *Node) {
	conn, err := follower.ConnPool.Acquire(context.Background())
	if err != nil {
		log.Printf("[ERROR] failed to acquire connection: %v", err)
	}
	defer conn.Release()
	subname := fmt.Sprintf("%s_%s_%d", follower.ID, leader.ID, time.Now().Nanosecond())
	_, _, dbname, user, password := util.DSNParse(follower.DSN)
	_, err = conn.Exec(context.Background(), query.CreateSubscriptionQuery(subname, leader.InternalHost, leader.InternalPort, dbname, user, password, leader.ID))
	if err != nil {
		log.Printf("[WARNING] failed to create subscription: %v", err)
	}
	log.Printf("[INFO] Subscription %s created for node %s", subname, follower.ID)
}
