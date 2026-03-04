package trnl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strings"
	"testing"
)

func TestRoutingAndHandling(t *testing.T) {
	r := Default()
	pUId := "/api/users/7878"
	pNoId := "/api/users"

	pMap := make(map[string]string)
	pMap["id"] = "7878"

	createBody := `{"name":"John Wick","user_id":"7878","email":"jh@example.com","password":"passingBy"}`
	updateBody := `{"name":"Jonathan Wick"}`
	getBody := `[{"name":"John Doe","user_id":"1234","email":"jd@example.com","password":"pass1234"},{"name":"Jane Doe","user_id":"5678","email":"jd1@example.com","password":"pass5678"},{"name":"Bob Dylan","user_id":"9012","email":"bd@example.com","password":"password1"},{"name":"Alice Jones","user_id":"3456","email":"aj@example.com","password":"word0987"},{"name":"Decagon Ten","user_id":"7890","email":"dt@example.com","password":"worded22"},{"name":"John Wick","user_id":"7878","email":"jh@example.com","password":"passingBy"}]`
	delBody := `[{"name":"John Doe","user_id":"1234","email":"jd@example.com","password":"pass1234"},{"name":"Jane Doe","user_id":"5678","email":"jd1@example.com","password":"pass5678"},{"name":"Bob Dylan","user_id":"9012","email":"bd@example.com","password":"password1"},{"name":"Alice Jones","user_id":"3456","email":"aj@example.com","password":"word0987"},{"name":"Decagon Ten","user_id":"7890","email":"dt@example.com","password":"worded22"}]`

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
		name       string
		req        *Request
		wantStatus int
		wantBody   string
	}{
		{
			name: "create user",
			req: &Request{
				Header: RequestHeader{Method: "POST", Path: pNoId},
				Body:   cBuf,
			},
			wantStatus: StatusOK,
			wantBody:   createBody,
		},
		{
			name: "get users",
			req: &Request{
				Header: RequestHeader{Method: "GET", Path: pNoId},
			},
			wantStatus: StatusOK,
			wantBody:   getBody,
		},
		{
			name: "get user by id",
			req: &Request{
				Header: RequestHeader{Method: "GET", Path: pUId},
			},
			wantStatus: StatusOK,
			wantBody:   createBody,
		},
		{
			name: "update user by id",
			req: &Request{
				Header: RequestHeader{Method: "PUT", Path: pUId},
				Body:   uBuf,
			},
			wantStatus: StatusOK,
			wantBody:   `{"name":"Jonathan Wick","user_id":"7878","email":"jh@example.com","password":"passingBy"}`,
		},
		{
			name: "delete user by id",
			req: &Request{
				Header: RequestHeader{Method: "DELETE", Path: pUId},
			},
			wantStatus: StatusOK,
			wantBody:   delBody,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			res := &response{
				writer: bufio.NewWriter(&buf),
				header: make(Header),
			}

			// Serve request
			r.ServeHTTP(res, tt.req)
			res.writer.Flush()

			// Check status code
			if res.status != tt.wantStatus {
				t.Errorf("status incorrect: got %v, want %v", res.status, tt.wantStatus)
				return
			}

			// Check response body
			gotBody := buf.String()
			parts := strings.SplitN(gotBody, "\r\n\r\n", 2)
			if len(parts) < 2 {
				t.Errorf("response format incorrect: got %v", gotBody)
				return
			}

			body := parts[1]
			if body != tt.wantBody {
				t.Errorf("body incorrect: got %v, want %v", body, tt.wantBody)
				return
			}
		})
	}
}

type mockUser struct {
	Name     string `json:"name"`
	UserId   string `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var db = []mockUser{
	mockUser{"John Doe", "1234", "jd@example.com", "pass1234"},
	mockUser{"Jane Doe", "5678", "jd1@example.com", "pass5678"},
	mockUser{"Bob Dylan", "9012", "bd@example.com", "password1"},
	mockUser{"Alice Jones", "3456", "aj@example.com", "word0987"},
	mockUser{"Decagon Ten", "7890", "dt@example.com", "worded22"},
}

func createUser(w ResponseWriter, r *Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Read error: %v", err)
		w.WriteHeader(StatusInternalServerError)
		return
	}

	dat := mockUser{}
	if err = json.Unmarshal(b, &dat); err != nil {
		log.Printf("unmarshal error: %v", err)
		w.WriteHeader(StatusBadRequest)
		return
	}

	usr := mockUser{Name: dat.Name, UserId: dat.UserId, Email: dat.Email, Password: dat.Password}
	db = append(db, usr)

	usrB, err := json.Marshal(usr)
	if err != nil {
		log.Printf("Marshalling error: %v", err)
		w.WriteHeader(StatusInternalServerError)
		return
	}

	w.WriteHeader(StatusOK)
	w.Write(usrB)
}

func getUsers(w ResponseWriter, r *Request) {
	users, err := json.Marshal(db)
	if err != nil {
		log.Printf("Marshalling error: %v", err)
		w.WriteHeader(StatusInternalServerError)
		return
	}

	w.WriteHeader(StatusOK)
	w.Write(users)
}

func getUserById(w ResponseWriter, r *Request) {
	userId, err := r.Params("id")
	if err != nil {
		log.Printf("Error retrieving parameter: %v", err)
		w.WriteHeader(StatusBadRequest)
		return
	}
	user := mockUser{}

	for _, usr := range db {
		if usr.UserId == userId {
			user = usr
			break
		}
	}

	usrB, err := json.Marshal(user)
	if err != nil {
		log.Printf("Marshalling error: %v", err)
		w.WriteHeader(StatusInternalServerError)
		return
	}

	w.WriteHeader(StatusOK)
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
		w.WriteHeader(StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(b, &prm); err != nil {
		log.Printf("Unmarshal error: %v", err)
		w.WriteHeader(StatusInternalServerError)
		return
	}

	userId, err := r.Params("id")
	if err != nil {
		log.Printf("Error retrieving parameter: %v", err)
		w.WriteHeader(StatusInternalServerError)
		return
	}
	user := mockUser{}

	for i := range db {
		if db[i].UserId == userId {
			db[i].Name = prm.Name
			user = db[i]
			break
		}
	}

	dat, err := json.Marshal(user)
	if err != nil {
		log.Printf("Marshalling error: %v", err)
		w.WriteHeader(StatusInternalServerError)
		return
	}

	w.WriteHeader(StatusOK)
	w.Write(dat)
}

func deleteUserById(w ResponseWriter, r *Request) {
	userId, err := r.Params("id")
	if err != nil {
		log.Printf("Error retrieving parameter: %v", err)
		w.WriteHeader(StatusInternalServerError)
		return
	}
	newUsers := []mockUser{}

	for _, usr := range db {
		if usr.UserId != userId {
			newUsers = append(newUsers, usr)
		}
	}
	db = newUsers

	users, err := json.Marshal(db)
	if err != nil {
		log.Printf("Marshalling error: %v", err)
		w.WriteHeader(StatusInternalServerError)
		return
	}

	w.WriteHeader(StatusOK)
	w.Write(users)
}
