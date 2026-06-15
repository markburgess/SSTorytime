// Package sstime provides functionality for working with SSTime data structures.

package sstime

import (
	"fmt"
)

// GetNodeOrbit returns the orbit of a node in the SSTime graph.
func GetNodeOrbit(node *Node) *Orbit {
	// ... existing code ...
	// Fix: Ensure the last item in the path search has a context.
	if len(path) > 0 {
		lastItem := path[len(path)-1]
		if lastItem.Context == nil {
			lastItem.Context = &Context{ /* default context */ }
		}
	}
	return orbit
}

// GetEntireCone returns the entire cone of a node in the SSTime graph.
func GetEntireCone(node *Node) *Cone {
	// ... existing code ...
	// Fix: Ensure the last item in the path search has a context.
	if len(cone.Path) > 0 {
		lastItem := cone.Path[len(cone.Path)-1]
		if lastItem.Context == nil {
			lastItem.Context = &Context{ /* default context */ }
		}
	}
	return cone
}