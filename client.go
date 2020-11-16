package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"othello/pkg/ws"
)

var user string

func init() {
	flag.StringVar(&user, "user", "pdf", "指定连接人")
}

func main() {
	flag.Parse()
	u := url.URL{Scheme: "ws", Host: "othello.pengdafu.ren:3333", Path: "/ws", RawQuery: "name=" + user}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	errs := make(chan error)

	go func() {
		for {
			msg := &ws.Message{}
			err := c.ReadJSON(msg)
			if err != nil {
				errs <- err
				log.Println("read:", err)
				return
			}
			if msg.Type != ACTION {
				continue
			}
			p := getPoint(msg.Board, msg.Color)
			msg.Point = p
			sendAnswer(c, errs, msg)
		}
	}()

	log.Fatal("Exited: ", <-errs)
}

func sendAnswer(c *websocket.Conn, ch chan<- error, msg *ws.Message) {
	err := c.WriteJSON(msg)
	if err != nil {
		ch <- err
	}
}

func getPoint(board [8][8]int8, color int8) ws.Point {

	for x, int8s := range board {
		for y, v := range int8s {
			if v != GRAY {
				continue
			}
			p := ws.Point{X: int8(x), Y: int8(y)}
			if findUp(board, p, color) {
				return p
			} else if findDown(board, p, color) {
				return p
			} else if findLeft(board, p, color) {
				return p
			} else if findRight(board, p, color) {
				return p
			} else if findLeftUp(board, p, color) {
				return p
			} else if findLeftDown(board, p, color) {
				return p
			} else if findRightUp(board, p, color) {
				return p
			} else if findRightDown(board, p, color) {
				return p
			}
		}
	}
	return ws.Point{}
}

func findUp(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.X-1 >= MIN {
		p.X = p.X - 1
		if board[p.X][p.Y] == -1*currentColor { // 是对手颜色继续判断
			v = true
			continue
		}
		if board[p.X][p.Y] == GRAY {
			return false
		}
		return v
	}
	return false
}

func findDown(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.X+1 <= MAX {
		p.X = p.X + 1
		if board[p.X][p.Y] == -1*currentColor { // 是对手颜色继续判断
			v = true
			continue
		}
		if board[p.X][p.Y] == GRAY {
			return false
		}
		return v
	}
	return false
}

func findLeft(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.Y-1 >= MIN {
		p.Y = p.Y- 1
		if board[p.X][p.Y] == -1*currentColor { // 是对手颜色继续判断
			v = true
			continue
		}
		if board[p.X][p.Y] == GRAY {
			return false
		}
		return v
	}
	return false
}

func findRight(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.Y+1 <= MAX {
		p.Y = p.Y + 1
		if board[p.X][p.Y] == -1*currentColor { // 是对手颜色继续判断
			v = true
			continue
		}
		if board[p.X][p.Y] == GRAY {
			return false
		}
		return v
	}
	return false
}

func findLeftUp(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.X-1 >= MIN && p.Y-1 >= MIN {
		p.X = p.X - 1
		p.Y = p.Y - 1
		if board[p.X][p.Y] == -1 * currentColor { // 是对手颜色继续判断
			v = true
			continue
		}
		if board[p.X][p.Y] == GRAY {
			return false
		}
		return v
	}
	return false
}

func findLeftDown(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.Y-1 >= MIN && p.X+1 <= MAX {
		p.Y = p.Y - 1
		p.X = p.X + 1
		if board[p.X][p.Y] == -1 * currentColor { // 是对手颜色继续判断
			v = true
			continue
		}
		if board[p.X][p.Y] == GRAY {
			return false
		}
		return v
	}
	return false
}

func findRightUp(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.Y+1 <= MAX && p.X-1 >= MIN {
		p.Y = p.Y + 1
		p.X = p.X - 1
		if board[p.X][p.Y] == -1 * currentColor { // 是对手颜色继续判断
			v = true
			continue
		}
		if board[p.X][p.Y] == GRAY {
			return false
		}
		return v
	}
	return false
}

func findRightDown(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.X+1 <= MAX && p.Y+1 <= MAX {
		p.X = p.X + 1
		p.Y = p.Y + 1
		if board[p.X][p.Y] == -1 * currentColor { // 是对手颜色继续判断
			v = true
			continue
		}
		if board[p.X][p.Y] == GRAY {
			return false
		}
		return v
	}
	return false
}
const (
	BLACK  = 1
	WHITE  = -1
	GRAY   = 0
	ACTION = "ACTION"
	MAX    = 7
	MIN    = 0
)
