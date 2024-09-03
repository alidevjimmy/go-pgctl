package node

type Rule int

const (
	Leader   Rule = 1
	Follower Rule = 2
)

type Node struct {
	Conn pg.conn
	Rule Rule
}
