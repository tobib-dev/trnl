package trnl

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"reflect"
	"strconv"
	"testing"
)

func TestAddAndMatchRoute(t *testing.T) {
	r := Default()
	pUId := "/api/users/7878"
	pNoId := "/api/users"

	pMap := make(map[string]string)
	pMap["id"] = "7878"

	createBody := `{"Name": "John Wick", "UserId": "7878", "Email": "jh@example.com", "Password": "passingBy"}`
	updateBody := `{"Name": "Jonathan Wick"}`

	cBytes := []byte(createBody)
	cBuf := bytes.NewBuffer(cBytes)
	uBytes := []byte(updateBody)
	uBuf := bytes.NewBuffer(uBytes)

	r.addRoute("GET", "/api/users", getUsers)
	r.addRoute("POST", "/api/users", createUser)
	r.addRoute("GET", "/api/users/:id", getUserById)

	r.addRoute("PUT", "/api/users/:id", updateUserById)
	r.addRoute("DELETE", "/api/users/:id", deleteUserById)

	tests := []struct {
		name string
		req  *Request
		want routeMatch
	}{
		{"create user", &Request{Header: RequestHeader{Method: "POST", Path: pNoId}, Body: cBuf}, routeMatch{HandlerFunc(createUser), nil, false, true}},
		{"get users", &Request{Header: RequestHeader{Method: "GET", Path: pNoId}}, routeMatch{HandlerFunc(getUsers), nil, false, true}},
		{"get user by id", &Request{Header: RequestHeader{Method: "GET", Path: pUId}}, routeMatch{HandlerFunc(getUserById), pMap, true, true}},
		{"update user by id", &Request{Header: RequestHeader{Method: "PUT", Path: pUId}, Body: uBuf}, routeMatch{HandlerFunc(updateUserById), pMap, true, true}},
		{"delete user by id", &Request{Header: RequestHeader{Method: "DELETE", Path: pUId}}, routeMatch{HandlerFunc(deleteUserById), pMap, true, true}},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			m := r.matchRoute(tt.req.Header.Method, tt.req.Header.Path)
			if reflect.ValueOf(m.handler).Pointer() != reflect.ValueOf(tt.want.handler).Pointer() {
				t.Errorf("handler incorrect: got %v, want %v", m.handler, tt.want.handler)
				return
			}

			// Normalize params maps before comparison
			if m.params == nil {
				m.params = make(map[string]string)
			}
			if tt.want.params == nil {
				tt.want.params = make(map[string]string)
			}

			if !reflect.DeepEqual(m.params, tt.want.params) {
				t.Errorf("invalid params: got %v, want %v", m.params, tt.want.params)
				return
			}

			if m.hasParams != tt.want.hasParams {
				t.Errorf("invalid route type: got %v, want %v", m.hasParams, tt.want.hasParams)
				return
			}

			if m.pathExists != tt.want.pathExists {
				t.Errorf("invalid path exists value: got %v, want %v", m.pathExists, tt.want.pathExists)
				return
			}
		})
	}
}

type mockUser struct {
	Name     string `json:"name"`
	UserId   int    `json:"userId"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var db = []mockUser{
	mockUser{"John Doe", 1234, "jd@example.com", "pass1234"},
	mockUser{"Jane Doe", 5678, "jd1@example.com", "pass5678"},
	mockUser{"Bob Dylan", 9012, "bd@example.com", "password1"},
	mockUser{"Alice Jones", 3456, "aj@example.com", "word0987"},
	mockUser{"Decagon Ten", 7890, "dt@example.com", "worded22"},
}

func createUser(w ResponseWriter, r *Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	dat := mockUser{}
	if err = json.Unmarshal(b, &dat); err != nil {
		log.Printf("unmarshal error: %v", err)
		return
	}

	usr := mockUser{Name: dat.Name, UserId: dat.UserId, Email: dat.Email, Password: dat.Password}
	db = append(db, usr)

	usrB, err := json.Marshal(usr)
	if err != nil {
		log.Printf("Marshalling error: %v", err)
	}
	w.Write(usrB)
}

func getUsers(w ResponseWriter, r *Request) {
	users, err := json.Marshal(db)
	if err != nil {
		log.Printf("Marshalling error: %v", err)
	}

	w.Write(users)
}

func getUserById(w ResponseWriter, r *Request) {
	userId := r.Params["id"]
	i64, err := strconv.ParseInt(userId, 10, 0)
	if err != nil {
		log.Printf("Int parse error: %v", err)
		return
	}
	usrId := int(i64)
	user := mockUser{}

	for _, usr := range db {
		if usr.UserId == usrId {
			user = usr
			break
		}
	}

	usrB, err := json.Marshal(user)
	if err != nil {
		log.Printf("Marshalling error: %v", err)
	}

	w.Write(usrB)
}

func updateUserById(w ResponseWriter, r *Request) {
	type param struct {
		Name string `json:"name"`
	}

	prm := param{}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Read request body error: %v", err)
		return
	}
	if err = json.Unmarshal(b, &prm); err != nil {
		log.Printf("Unmarshal error: %v", err)
		return
	}

	userId := r.Params["id"]
	i64, err := strconv.ParseInt(userId, 10, 0)
	if err != nil {
		log.Printf("Int parse error: %v", err)
		return
	}
	usrId := int(i64)
	user := mockUser{}

	for i := range db {
		if db[i].UserId == usrId {
			db[i].Name = prm.Name
			user = db[i]
			break
		}
	}

	dat, err := json.Marshal(user)
	if err != nil {
		log.Printf("Marshalling error: %v", err)
	}

	w.Write(dat)
}

func deleteUserById(w ResponseWriter, r *Request) {
	userId := r.Params["id"]
	i64, err := strconv.ParseInt(userId, 10, 0)
	if err != nil {
		log.Printf("Int parse error: %v", err)
		return
	}
	usrId := int(i64)
	newUsers := []mockUser{}

	for _, usr := range db {
		if usr.UserId != usrId {
			newUsers = append(newUsers, usr)
		}
	}
	db = newUsers

	users, err := json.Marshal(db)
	if err != nil {
		log.Printf("Marshalling error: %v", err)
	}

	w.Write(users)
}
