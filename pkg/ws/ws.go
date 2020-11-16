package ws

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	CLOSED = 1 << iota
	CONNECTING
)

type User struct {
	UserName string
	Conn     *websocket.Conn
	Writer   chan *Message
	State    int32
}

type Othello struct {
	sync.Mutex
	Users map[string]*User
}

type Message struct {
	Type  string     `json:"type"`
	Name  string     `json:"name"`
	Key   string     `json:"key"`
	Point Point      `json:"point"`
	Board [8][8]int8 `json:"board"`
	Err   string     `json:"err"`
	Color int8       `json:"color"`
}

type Point struct {
	X int8 `json:"x"`
	Y int8 `json:"y"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(resquest *http.Request) bool {
		return true
	},
}

func (o *Othello) Client(ctx *gin.Context) {
	log.Println("一个新连接~")
	c, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println("Socket connect error", err)
		return
	}
	Name := ctx.Query("name")
	if Name == "" {
		log.Println("需要制定一个身份")
		return
	}
	go o.Register(c, Name)
}

func (o *Othello) Register(conn *websocket.Conn, Name string) {
	o.Lock()
	defer o.Unlock()
	if o.Users == nil {
		o.Users = map[string]*User{}
	}
	if _, ok := o.Users[Name]; !ok {
		o.Users[Name] = &User{
			UserName: Name,
			Conn:     conn,
			Writer:   make(chan *Message),
			State:    CONNECTING,
		}
	}
	user := o.Users[Name]
	if user.Conn == nil || user.Conn != conn {
		user.Closed()
		user.Conn = conn
		user.Writer = make(chan *Message)
	}
	go user.readMessage()
	go user.writeMessage()
	user.Writer <- &Message{Name: Name, Type: "Confirm"}
	log.Println("连接已建立~")
}

func (u *User) writeMessage() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
		u.Closed()
	}()
	for {
		msg := <-u.Writer
		log.Println("准备发送一个消息", msg)
		if err := u.Conn.WriteJSON(msg); err != nil {
			log.Println("Send msg error", err)
			break
		}
		log.Println("成功发送一个消息", msg)
	}
}

type MessageBus struct {
	sync.RWMutex
	Bus map[string]chan *Message
}

var msgBus *MessageBus

func init() {
	msgBus = &MessageBus{
		Bus: map[string]chan *Message{},
	}
}

func RegisterBus(key string) {
	msgBus.Lock()
	defer msgBus.Unlock()
	msgBus.Bus[key] = make(chan *Message, 1)
}

func sendMsgToBus(msg *Message) {
	ch, ok := msgBus.Bus[msg.Key]
	if !ok {
		return
	}
	ch <- msg
}

func (u *User) GetMessage(key string) *Message {
	msgBus.RLock()
	defer msgBus.RUnlock()
	ch, ok := msgBus.Bus[key]
	if !ok {
		return nil
	}
	select {
	case m := <-ch:
		return m
	case <-time.After(3 * time.Second):
		return &Message{
			Key: key,
			Err: "获取信息超时",
		}
	}
}

func (u *User) readMessage() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
		u.Closed()
	}()
	for {
		msg := &Message{}
		err := u.Conn.ReadJSON(msg)
		if err != nil {
			log.Printf("[Error ReadMessage]: %v 断开连接 -> %v", u.UserName, err)
			break
		}
		if msg.Key == "" {
			continue
		}
		sendMsgToBus(msg)
	}
}

func (u *User) Closed() {
	u.Conn.Close()
	atomic.StoreInt32(&u.State, CLOSED)
}

func (u *User) Checked() bool {
	return u.State == CONNECTING
}
