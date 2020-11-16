package play

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"othello/model"
	"othello/pkg/ws"
	"time"
)

const (
	BLACK  = 1
	WHITE  = -1
	GRAY   = 0
	ACTION = "ACTION"
)

func Gaming(ctx *gin.Context, o *ws.Othello) {
	first := ctx.Param("first")
	back := ctx.Param("back")
	if first == "" || back == "" {
		ctx.JSON(400, gin.H{
			"msg":  "参赛者不能为空",
			"code": 400,
		})
		return
	}
	go playGame(o.Users[first], o.Users[back])
	ctx.JSON(200, gin.H{
		"msg":  "比赛中",
		"code": 200,
	})
}

func GameOver(ctx *gin.Context) {
	gr := &model.GameResult{}
	first := ctx.Param("first")
	back := ctx.Param("back")
	if first == "" || back == "" {
		ctx.JSON(400, gin.H{
			"msg":  "参赛者不能为空",
			"code": 400,
		})
		return
	}
	_, rows := gr.FindByFirstAndBack(first, back)
	data := &[]gameData{}
	if rows > 0 {
		_ = json.Unmarshal([]byte(gr.Data), data)
		ctx.JSON(200, gin.H{
			"msg":         gr,
			"code":        200,
			"DataProcess": data,
		})
		return
	}
	ctx.JSON(400, gin.H{
		"msg":  "还没有比赛~",
		"code": 400,
	})
}

// newBoard 白棋先手
func newBoard() [8][8]int8 {
	board := [8][8]int8{}
	board[3][3] = WHITE
	board[4][3] = BLACK
	board[3][4] = BLACK
	board[4][4] = WHITE
	return board
}

type gameData struct {
	Point ws.Point `json:"point"`
	Gamer string   `json:"gamer"`
	Color int8     `json:"color"`
}

func playGame(first, back *ws.User) {
	gr := model.GameResult{First: first.UserName, Back: back.UserName}
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			gr.Reason = "比赛异常结束"
			_ = gr.Insert()
		}
	}()
	board := newBoard()
	gd := []gameData{}
	key := fmt.Sprintf("%v", time.Now().UnixNano())
	ws.RegisterBus(key)
	funcDo := func(currentColor, enemyColor int8, u *ws.User, win int8, enemyUser *ws.User) bool {
		for {
			u.Writer <- &ws.Message{Key: key, Type: ACTION, Name: u.UserName, Board: board, Color: currentColor}
			msg := u.GetMessage(key)
			if msg.Err != "" {
				gr.Reason = u.UserName + msg.Err
				gr.Win = win * -1
				gr.WinChessPieces = getWinChessPieces(board)
				d, _ := json.Marshal(&gd)
				gr.Data = string(d)
				_ = gr.Insert()
				return true
			}

			p := msg.Point
			gd = append(gd, gameData{Color: currentColor, Gamer: u.UserName, Point: p})
			log.Println(u.UserName, "当前下子", p)
			if p.X == 7 && p.Y == 7 {
				log.Println("in")
			}
			// 检查落子是否有问题
			if !checkAndSet(&board, currentColor, p) { // 下子错误，直接判输
				gr.Reason = u.UserName + "错误落子点，对方直接赢"
				gr.Win = win * -1
				gr.WinChessPieces = getWinChessPieces(board)
				d, _ := json.Marshal(&gd)
				gr.Data = string(d)
				_ = gr.Insert()
				return true
			}
			winS := checkWin(board)
			if winS.IsWin {
				if winS.WinColor == 0 {
					gr.Reason = "平局"
				} else if winS.WinColor == currentColor {
					gr.Reason = u.UserName + "胜利"
					gr.Win = win
				} else {
					gr.Reason = enemyUser.UserName + "胜利"
					gr.Win = -1 * win
				}

				gr.WinChessPieces = winS.WinChessPieces
				d, _ := json.Marshal(&gd)
				gr.Data = string(d)
				log.Println("胜利时棋盘:", board)
				_ = gr.Insert()
				return true
			}
			if !check(board, enemyColor) {
				continue
			}

			break
		}
		return false
	}
	for {
		if funcDo(WHITE, BLACK, first, 1, back) {
			break
		}
		if funcDo(BLACK, WHITE, back, -1, first) {
			break
		}
	}
}

type WinStatus struct {
	IsWin          bool
	WinColor       int8
	WinChessPieces uint8
}

func checkWin(board [8][8]int8) WinStatus {
	blackNum := 0
	whiteNum := 0
	grayNum := 0
	for _, int8s := range board {
		for _, v := range int8s {
			if v == BLACK {
				blackNum++
			} else if v == WHITE {
				whiteNum++
			} else if v == GRAY {
				grayNum++
			}
		}
	}
	winStatus := WinStatus{}
	if blackNum == 0 || whiteNum == 0 || grayNum == 0 {
		winStatus.IsWin = true
		winStatus.WinChessPieces = uint8(math.Abs(float64(blackNum - whiteNum)))
		if blackNum > whiteNum {
			winStatus.WinColor = BLACK
		} else if blackNum == whiteNum {
			winStatus.WinColor = 0
		} else {
			winStatus.WinColor = WHITE
		}
	}
	return winStatus
}

func getWinChessPieces(board [8][8]int8) uint8 {
	blackNum := 0
	whiteNum := 0
	for _, int8s := range board {
		for _, i3 := range int8s {
			if i3 == WHITE {
				whiteNum++
			} else if i3 == BLACK {
				blackNum++
			}
		}
	}
	return uint8(math.Abs(float64(blackNum - whiteNum)))
}

var directive = map[string]ws.Point{
	"LEFT": ws.Point{
		X: 0,
		Y: -1,
	},
	"RIGHT": ws.Point{
		X: 0,
		Y: 1,
	}, "UP": ws.Point{
		X: -1,
		Y: 0,
	}, "DOWN": ws.Point{
		X: 1,
		Y: 0,
	}, "LEFTUP": ws.Point{
		X: -1,
		Y: -1,
	}, "LEFTDOWN": ws.Point{
		X: 1,
		Y: -1,
	}, "RIGHTUP": ws.Point{
		X: -1,
		Y: 1,
	}, "RIGHTDOWN": ws.Point{
		X: 1,
		Y: 1,
	},
}

func check(board [8][8]int8, Color int8) (changed bool) {
	for x, int8s := range board {
		for y, v := range int8s {
			point := ws.Point{X: int8(x), Y: int8(y)}
			if v == GRAY && (findUp(board, point, Color) ||
				findDown(board, point, Color) ||
				findLeft(board, point, Color) ||
				findRight(board, point, Color) ||
				findLeftDown(board, point, Color) ||
				findLeftUp(board, point, Color) ||
				findRightUp(board, point, Color) ||
				findRightDown(board, point, Color)) {
				return true
			}
		}
	}
	return false
}

func checkAndSet(board *[8][8]int8, Color int8, point ws.Point) (changed bool) {
	if board[point.X][point.Y] != GRAY {
		return false
	}
	if findUp(*board, point, Color) {
		changed = true
		set(board, point, Color, directive["UP"])
	}
	if findDown(*board, point, Color) {
		changed = true
		set(board, point, Color, directive["DOWN"])
	}
	if findLeft(*board, point, Color) {
		changed = true
		set(board, point, Color, directive["LEFT"])
	}
	if findRight(*board, point, Color) {
		changed = true
		set(board, point, Color, directive["RIGHT"])
	}
	if findLeftDown(*board, point, Color) {
		changed = true
		set(board, point, Color, directive["LEFTDOWN"])
	}
	if findLeftUp(*board, point, Color) {
		changed = true
		set(board, point, Color, directive["LEFTUP"])
	}
	if findRightUp(*board, point, Color) {
		changed = true
		set(board, point, Color, directive["RIGHTUP"])
	}
	if findRightDown(*board, point, Color) {
		changed = true
		set(board, point, Color, directive["RIGHTDOWN"])
	}
	if changed {
		board[point.X][point.Y] = Color
	}
	return
}

const (
	MIN = 0
	MAX = 7
)

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
		p.Y = p.Y - 1
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

func findLeftDown(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.Y-1 >= MIN && p.X+1 <= MAX {
		p.Y = p.Y - 1
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

func findRightUp(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.Y+1 <= MAX && p.X-1 >= MIN {
		p.Y = p.Y + 1
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

func findRightDown(board [8][8]int8, p ws.Point, currentColor int8) bool {
	v := false
	for p.X+1 <= MAX && p.Y+1 <= MAX {
		p.X = p.X + 1
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

func set(board *[8][8]int8, p ws.Point, Color int8, subp ws.Point) {
	p.Y = p.Y + subp.Y
	p.X = p.X + subp.X
	for board[p.X][p.Y] != Color {
		board[p.X][p.Y] = Color // 刚好填补空位
		p.Y = p.Y + subp.Y
		p.X = p.X + subp.X
	}
}
