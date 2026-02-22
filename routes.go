package trnl

/*
 * Serve request using method and route, respond with 404 if endpoint is not found
 */
func (m *Mux) ServeHTTP(w ResponseWriter, r *Request) {
	// match routes to get handler and params if endpoint is dynamic
	match := m.matchRoute(r.Header.Method, r.Header.Path)
	if match.handler == nil && !match.pathExists {
		// Not Found
		w.WriteHeader(StatusNotFound)
		return
	} else if match.handler == nil && match.pathExists {
		// Method is not allowed
		w.WriteHeader(StatusMethodNotAllowed)
		return
	}

	// If endpoint has parameters load params into the request
	if match.hasParams {
		for k, v := range match.params {
			r.Params[k] = v
		}
	}

	match.handler.ServeHTTP(w, r)
}

/*
 * Initializes default multiplexer and return mux address to user
 */
func Default() *Mux {
	return &Mux{
		root: &rNode{
			children: make(map[string]*rNode),
			handlers: make(map[string]Handler),
		},
	}
}

func (m *Mux) Get(path string, handler any) {
	m.addRoute("GET", path, handler)
}

func (m *Mux) Post(path string, handler any) {
	m.addRoute("POST", path, handler)
}

func (m *Mux) Put(path string, handler any) {
	m.addRoute("PUT", path, handler)
}

func (m *Mux) Delete(path string, handler any) {
	m.addRoute("DELETE", path, handler)
}
