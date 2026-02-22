package trnl

import (
	"fmt"
	"slices"
	"strings"
)

var validRoutes = []string{"GET", "POST", "PUT", "DELETE"}

type rNode struct {
	children     map[string]*rNode  // Map subpath string to children node
	handlers     map[string]Handler // Map HTTP methods to handler functions
	dynamicRoute bool
	param        string
	endOfPath    bool
}

type Mux struct {
	root *rNode
}

// Return value for matchRoute
type routeMatch struct {
	handler    Handler
	params     map[string]string
	hasParams  bool
	pathExists bool
}

/*
 * Registers user defined endpoint to the router trie
 */
func (m *Mux) addRoute(method, path string, handlr any) {
	if !slices.Contains(validRoutes, method) {
		panic("method must be one of the following: GET, POST, PUT, DELETE")
	}

	if path == "" || path[0] != '/' {
		panic("routes must start with '/'")
	}

	segments := filterEmpty(strings.Split(path, "/"))
	node := m.root

	for _, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			// Handle dynamic segment
			prm := seg[1:]

			// Create child node if it doesn't exist and set to dynamic
			child, ok := node.children[":"]
			if !ok {
				child = &rNode{
					dynamicRoute: true,
					param:        prm,
					children:     make(map[string]*rNode),
					handlers:     make(map[string]Handler),
				}

				node.children[":"] = child
			} else if child.param != prm && child.param != "" {
				msg := fmt.Sprintf("conflicting param names for dynamic segments: %s => %s - %s", seg, child.param, prm)
				panic(msg)
			}
			node = child
		} else {
			// Handle static segment
			_, ok := node.children[seg]
			if !ok {
				node.children[seg] = &rNode{
					children: make(map[string]*rNode),
					handlers: make(map[string]Handler),
				}
			}
			node = node.children[seg]
		}
	}

	// Check if method exist on path after adding subpaths
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

/*
 * Returns handler, parameters, and boolean value verifying if route is static or dynamic.
 * Will return nil handler if method is not allowed or no static nor dynamic routes are
 * found
 */
func (m *Mux) matchRoute(method, path string) routeMatch {
	node := m.root
	segments := filterEmpty(strings.Split(path, "/"))
	match := routeMatch{
		params:     make(map[string]string),
		hasParams:  false,
		pathExists: false,
	}

	for _, seg := range segments {
		// Direct match on a static route
		if child, ok := node.children[seg]; ok {
			node = child
			continue
		}

		// Dynamic routes
		if child, ok := node.children[":"]; ok {
			node = child
			if node.param != "" {
				match.params[node.param] = seg
				match.hasParams = true
			}
			continue
		}

		return match // returns nil handler if no static or dynamic routes
	}

	// Check if method is allowed for path
	handlr, ok := node.handlers[method]
	match.pathExists = true
	if !ok {
		return match // method not allowed
	}

	match.handler = handlr
	return match // return handler, params if found and boolean (true - has params, false - no params)
}

func filterEmpty(parts []string) []string {
	fPath := parts[:0]

	// users//:id ["users", "", ":id"] => ["users", ":id"]
	for _, p := range parts {
		if p != "" {
			fPath = append(fPath, p)
		}
	}

	return fPath
}
