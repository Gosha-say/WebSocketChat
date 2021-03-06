package main

import (
	"database/sql"
	b64 "encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
)

const html = "PCFET0NUWVBFIGh0bWw+CjxodG1sIGxhbmc9InJ1Ij4KPGhlYWQ+CiAgICA8bWV0YSBjaGFyc2V0PSJVVEYtOCIgLz4KICAgIDxtZXRhIG5hbWU9InZpZXdwb3J0IiBjb250ZW50PSJ3aWR0aD1kZXZpY2Utd2lkdGgsIGluaXRpYWwtc2NhbGU9MS4wIiAvPgogICAgPG1ldGEgaHR0cC1lcXVpdj0iWC1VQS1Db21wYXRpYmxlIiBjb250ZW50PSJpZT1lZGdlIiAvPgogICAgPHRpdGxlPkdvIFdlYlNvY2tldCBUdXRvcmlhbDwvdGl0bGU+CjwvaGVhZD4KPGJvZHk+CjxoMj5IZWxsbyBXb3JsZDwvaDI+CjxwIGlkPSJvdXRwdXQiPjwvcD4KPGxhYmVsIGZvcj0iaW5wdXQiPk1lc3NhZ2U6IDwvbGFiZWw+PGlucHV0IGlkPSJpbnB1dCIgdHlwZT0idGV4dCIgLz4KPGJ1dHRvbiBvbmNsaWNrPSJzZW5kKCkiPlNlbmQ8L2J1dHRvbj4KPHNjcmlwdD4KICAgIGxldCBzb2NrZXQgPSBuZXcgV2ViU29ja2V0KCJ3czovLzEyNy4wLjAuMTo4MDgwL3dzIik7CiAgICBsZXQgaW5wdXQgPSBkb2N1bWVudC5nZXRFbGVtZW50QnlJZCgiaW5wdXQiKTsKCiAgICBjb25zb2xlLmxvZygiQXR0ZW1wdGluZyBDb25uZWN0aW9uLi4uIik7CgovKiAgICBzb2NrZXQub25vcGVuID0gKCkgPT4gewogICAgICAgIGNvbnNvbGUubG9nKCJTdWNjZXNzZnVsbHkgQ29ubmVjdGVkIik7CiAgICAgICAgc29ja2V0LnNlbmQoIkhpIEZyb20gdGhlIENsaWVudCEiKQogICAgfTsKCiAgICBzb2NrZXQub25jbG9zZSA9IGV2ZW50ID0+IHsKICAgICAgICBjb25zb2xlLmxvZygiU29ja2V0IENsb3NlZCBDb25uZWN0aW9uOiAiLCBldmVudCk7CiAgICAgICAgc29ja2V0LnNlbmQoIkNsaWVudCBDbG9zZWQhIikKICAgIH07Ki8KCiAgICBzb2NrZXQub25lcnJvciA9IGVycm9yID0+IHsKICAgICAgICBjb25zb2xlLmxvZygiU29ja2V0IEVycm9yOiAiLCBlcnJvcik7CiAgICB9OwoKICAgIHNvY2tldC5vbm1lc3NhZ2UgPSBmdW5jdGlvbihldnQpIHsKICAgICAgICBsZXQgb3V0ID0gZG9jdW1lbnQuZ2V0RWxlbWVudEJ5SWQoJ291dHB1dCcpOwogICAgICAgIGNvbnNvbGUubG9nKGV2dCk7CiAgICAgICAgbGV0IG1zZyA9SlNPTi5wYXJzZShldnQuZGF0YSk7CiAgICAgICAgb3V0LmlubmVySFRNTCArPSBtc2cuYm9keSArICc8YnI+JzsKICAgIH07CgogICAgZnVuY3Rpb24gc2VuZCgpIHsKICAgICAgICBzb2NrZXQuc2VuZChpbnB1dC52YWx1ZSk7CiAgICAgICAgaW5wdXQudmFsdWUgPSAiIjsKICAgIH0KCjwvc2NyaXB0Pgo8L2JvZHk+CjwvaHRtbD4="

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
	}
}

func serveWs(pool *Pool, w http.ResponseWriter, r *http.Request) {
	fmt.Println("WebSocket Endpoint Hit")
	conn, err := Upgrade(w, r)
	if err != nil {
		_, err := fmt.Fprintf(w, "%+v\n", err)
		checkErr(err)
	}

	client := &Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read()
}

func setupRoutes() {
	pool := NewPool()
	go pool.Start()
	//tpl, err := ioutil.ReadFile("./tpl.html")
	//tpl64 := b64.StdEncoding.EncodeToString(tpl)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := b64.StdEncoding.DecodeString(html)
		_, err := fmt.Fprintf(w, string(body))
		checkErr(err)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(pool, w, r)
	})
}

func main() {
	fmt.Println("Chat App v0.02")
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		fmt.Print("Data dir is not exists, try to create")
		if os.Mkdir("./data", os.ModePerm) != nil {
			log.Fatal("Can not create data dir")
		}
		if _, err := os.Stat("./data"); os.IsNotExist(err) {
			log.Fatal("Dir create - fail")
		}
		fmt.Println(" - OK")
	}
	if _, err := os.Stat("./data/db.sqlite"); os.IsNotExist(err) {
		fmt.Print("DB file is not exists, try to create")
		_, err := os.Create("./data/db.sqlite")
		if err != nil {
			log.Fatal("Can not create DB file")
		}
		if _, err := os.Stat("./data/db.sqlite"); os.IsNotExist(err) {
			log.Fatal("File create - fail")
		}
		fmt.Println(" - OK")
	}
	db, err := sql.Open("sqlite3", "./data/db.sqlite")
	checkErr(err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS main.members
(
	id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
	name  VARCHAR(16), 
	time timestamp default (strftime('%s', 'now')),
	key varchar(12)
);`)

	checkErr(err)

	rows, err := db.Query("select name from main.members where name = 'Server'")
	checkErr(err)
	defer rows.Close()
	checkErr(rows.Err())
	if !rows.Next() {
		id := uuid.New()
		fmt.Println("Creating user 'server' with id " + id.String())
		result, err := db.Exec("INSERT INTO main.members (name, key) values ('Server', ?)", id.String())
		checkErr(err)
		fmt.Print("Rows: ")
		fmt.Println(result.RowsAffected())
	} else {
		fmt.Println("User 'Server' exists")
	}

	fmt.Print("Try to create table 'messages'")
	_, err = db.Exec(`create table if not exists messages
(
	id integer primary key autoincrement unique,
	time timestamp default (strftime('%s', 'now')),
	member integer not null,
	type integer default 1 not null,
	body varchar(255) not null,
	other text
);`)
	checkErr(err)
	fmt.Println(" - OK")
	_, err = db.Exec("INSERT INTO main.messages (member, type, body) values (1, 1, 'Begin...')")
	checkErr(err)
	checkErr(db.Close())

	setupRoutes()
	checkErr(http.ListenAndServe(":8080", nil))
}

func (pool *Pool) Start() {
	log.Println("Start...")
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client := range pool.Clients {
				fmt.Println(client)
				checkErr(client.Conn.WriteJSON(Message{Type: 1, Body: "New User Joined..."}))
			}
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			for client := range pool.Clients {
				checkErr(client.Conn.WriteJSON(Message{Type: 1, Body: "User Disconnected..."}))
			}
			break
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in Pool")
			for client := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		checkErr(c.Conn.Close())
	}()

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		message := Message{Type: messageType, Body: string(p)}
		c.Pool.Broadcast <- message
		fmt.Printf("Message Received: %+v\n", message)
	}
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return conn, nil
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
