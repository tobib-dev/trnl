package trnl

import (
	"fmt"
	"slices"
	"strings"
)

var validRoutes = []string{"GET", "POST", "UPDATE", "DELETE"}

type RadixNode struct {
	children       map[string]*RadixNode // Map parent node to children node
	handlers       map[string]Handler    // Map HTTP methods to handler functions
	isDynamicRoute bool
	param          string
	endOfPath      bool
}

type TrieRouter struct {
	root *RadixNode
}

func (t *TrieRouter) addRoute(method, path string, handlr any) {
	if !slices.Contains(validRoutes, method) {
		panic("method must be one of the following: GET, POST, UPDATE, DELETE")
	}

	if path == "" || path[0] != '/' {
		panic("routes must start with '/'")
	}

	segments := filterEmpty(strings.Split(path, "/"))
	node := t.root

	for _, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			// Handle dynamic segment
			paramName := seg[1:]

			// Create child node and set to dynamic
			child, ok := node.children[":"]
			if !ok {
				child = &RadixNode{
					isDynamicRoute: true,
					param:          paramName,
					children:       make(map[string]*RadixNode),
					handlers:       make(map[string]Handler),
				}

				node.children[":"] = child
			} else if child.param != paramName && child.param != "" {
				msg := fmt.Sprintf("conflicting param names for dynamic segments: %s => %s - %s", seg, child.param, paramName)
				panic(msg)
			}
			node = child
		} else {
			// Handle static segment
			_, ok := node.children[seg]
			if !ok {
				node.children[seg] = &RadixNode{
					children: make(map[string]*RadixNode),
					handlers: make(map[string]Handler),
				}
			}
			node = node.children[seg]
		}
	}
	if _, exists := node.handlers[method]; exists {
		panic("already registered route: " + method + " " + path)
	}

	switch h := handlr.(type) {
	case Handler:
		node.handlers[method] = h
	case func(ResponseWriter, *Request):
		node.handlers[method] = HandlerFunc(h)
	default:
		panic("Invalid handler type, handler must be Handler interface or func(ResponseWriter, *Request)")
	}

	node.endOfPath = true // Mark that at least one handler exists at this path
}

func filterEmpty(parts []string) []string {
	filteredPath := parts[:0]

	for _, p := range parts {
		if p != "" {
			filteredPath = append(filteredPath, p)
		}
	}

	return filteredPath
}

func DefaultTrieRouter() *TrieRouter {
	return &TrieRouter{
		root: &RadixNode{
			children: make(map[string]*RadixNode),
			handlers: make(map[string]Handler),
		},
	}
}
