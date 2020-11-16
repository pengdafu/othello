package play

import (
  "github.com/gin-gonic/gin"
  "othello/model"
  "othello/pkg/ws"
)

func Show(ctx *gin.Context) {
  first := ctx.Query("first")
  back := ctx.Query("back")
  gr := &model.GameResult{}
  _, rows := gr.FindByFirstAndBack(first, back)
  if rows == 0 {
    ctx.JSON(400, gin.H{
      "code": 400,
      "msg": "比赛无结果",
    })
    return
  }
  ctx.HTML(200, "show.html", gin.H{
    "first": first,
    "back": back,
    "res": gr,
  })
}

func Index(ctx *gin.Context, o *ws.Othello) {
  gr := &model.GameResult{}
  res, _ := gr.GetRank()
  allRes, _ := gr.GetAllGameResult()
  xx := []model.GameResult{}
  for first, _ := range o.Users {
    for back, _ := range o.Users {
      if first == back {
        continue
      }
      gr0 := model.GameResult{First: first, Back: back, Reason: "未进行比赛", Win: 2}
      for _, gameResult := range allRes {
        if gameResult.First == first && gameResult.Back == back {
          gr0 = gameResult
          break
        }
      }
      xx = append(xx, gr0)
    }
  }
  ctx.HTML(200, "index.html", gin.H{
    "rank": res,
    "users": xx,
  })
}