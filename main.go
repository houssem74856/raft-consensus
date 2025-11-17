package main

import (
	"fmt"
	"simple-go-consensus/node"
)

func main() {
	ids := []string{"n1", "n2", "n3"}

	for _, id := range ids {
		go func(id string) {
			n := node.NewNode(id)
			n.Start()
		}(id)
	}

	fmt.Println("nodes started")
	select {} // block forever to keep program running while nodes operate
}
