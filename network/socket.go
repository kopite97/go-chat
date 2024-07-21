package network

import (
	. "chat_server_golang/types"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type message struct {
	Name    string
	Message string
	Time    int64
}

var upgrader = &websocket.Upgrader{ReadBufferSize: SocketBufferSize, WriteBufferSize: MessageBufferSize, CheckOrigin: func(r *http.Request) bool { return true }}

// 채팅방에 대한 값
type Room struct {
	Foward chan *message // 수신되는 메시지를 보관하는 값
	// 들어오는 메시지를 다른 클라이언트들에게 전송을 함

	Join  chan *client // Socket이 연결되는 경우에 작동
	Leave chan *client // Socket이 끊어지는 경우에 대해서 작동

	Clients map[*client]bool // 현재 방에 있는 Client 정보를 저장
}

type client struct {
	Send   chan *message
	Room   *Room
	Name   string
	Socket *websocket.Conn
}

func NewRoom() *Room {
	return &Room{
		Foward:  make(chan *message),
		Join:    make(chan *client),
		Leave:   make(chan *client),
		Clients: make(map[*client]bool),
	}
}

func (c *client) Read() {

	defer c.Socket.Close()
	// 클라이언트가 들어오는 메시지를 읽는 함수
	for {
		var msg *message
		err := c.Socket.ReadJSON(&msg)
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				break
			} else {
				panic(err)
			}
		} else {
			log.Println("READ : ", msg, "client", c.Name)
			log.Println()
			msg.Time = time.Now().Unix()
			msg.Name = c.Name
		}
		c.Room.Foward <- msg
	}
}

func (c *client) Write() {

	defer c.Socket.Close()

	// 클라이언트가 메시지를 전송하는 함수

	for msg := range c.Send {
		log.Println("WRITE :", msg, "client", c.Name)
		log.Println()
		err := c.Socket.WriteJSON(msg)
		if err != nil {
			panic(err)
		}
	}
}

func (r *Room) RunInit() {
	// Room에 있는 모든 채널값들을 받는 역할
	for {
		select {
		case client := <-r.Join:
			r.Clients[client] = true
		case client := <-r.Leave:
			r.Clients[client] = false
			close(client.Send)
			delete(r.Clients, client)
		case msg := <-r.Foward:
			for client := range r.Clients {
				client.Send <- msg
			}
		}
	}
}

// gin.Context를 사용해야 함 하지않으면 API를 통해 연결이 불가
func (r *Room) SocketServe(c *gin.Context) {

	socket, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}

	userCookie, err := c.Request.Cookie("auth")
	if err != nil {
		panic(err)
	}

	client := &client{
		Socket: socket,
		Send:   make(chan *message, MessageBufferSize),
		Room:   r,
		Name:   userCookie.Value,
	}

	r.Join <- client

	// defer를 사용하면 쉽게 함수 마지막에 코드를 작성할 수 있다.
	// 함수 블록을 나가기 직전 실행됨
	defer func() { r.Leave <- client }()

	go client.Write()

	client.Read()
}
