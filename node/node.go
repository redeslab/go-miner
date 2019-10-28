package node

import "sync"

var (
	instance *Node = nil
	once     sync.Once
)

type Node struct {
}

func Inst() *Node {
	once.Do(func() {
		instance = newNode()
	})
	return instance
}

func newNode() *Node {
	n := &Node{}
	return n
}

func Mining() {

}
