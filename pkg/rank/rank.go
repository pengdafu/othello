package rank

import (
  "github.com/gin-gonic/gin"
  "othello/model"
  "othello/pkg/ws"
)

func GetRank(ctx *gin.Context) {
  gr := &model.GameResult{}

  res, _ := gr.GetRank()
  ctx.JSON(200, gin.H{
    "res": res,
    "code": 200,
  })
}

func GetUsers(ctx *gin.Context, o *ws.Othello) {
  users := []string{}
  for name, _ := range o.Users {
    users = append(users, name)
  }
  ctx.JSON(200, gin.H{
    "code": 200,
    "users": users,
  })
}