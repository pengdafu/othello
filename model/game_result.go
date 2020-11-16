package model

type GameResult struct {
	First          string
	Back           string
	Win            int8 // 1 赢了 -1 对手赢了
	Data           string
	WinChessPieces uint8
	Reason         string
}

func (*GameResult) TableName() string {
	return "game_result"
}

func (r *GameResult) Insert() error {
	db.Where("first = ? and back = ?", r.First, r.Back).Delete(r)
	return db.Create(r).Error
}

func (r *GameResult) FindByFirstAndBack(first, back string) (err error, rows int64) {
	re := db.Where("first = ? AND back = ?", first, back).Find(r)
	return re.Error, re.RowsAffected
}

func (r *GameResult) GetRank() (res []Rank, err error) {
	err = db.Raw(`
SELECT name, SUM(score) score, SUM(WinChessPieces) WinChessPieces
FROM (
         SELECT first name, SUM(if(win = 1, 1, 0)) score, SUM(IF(win = 1, win_chess_pieces, 0)) WinChessPieces
         FROM game_result
         GROUP BY name
         union
         SELECT back name, SUM(if(win = -1, 1, 0)) score, SUM(IF(win = -1, win_chess_pieces, 0)) WinChessPieces
         FROM game_result
         GROUP BY name
     ) t
GROUP BY name
order by score desc, WinChessPieces desc
`).Scan(&res).Error
	if err != nil {
		return
	}
	rank := uint8(1)
	for i := 0; i < len(res); i++ {
		re := &res[i]
		re.Rank = rank
		rank++
	}
	return
}

func (*GameResult) GetAllGameResult()(res []GameResult, err error) {
	err = db.Raw("SELECT * FROM game_result").Scan(&res).Error
	return
}

type Rank struct {
	Name           string
	Rank           uint8
	Score          uint8
	WinChessPieces uint8
}
