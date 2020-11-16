package router

import (
	"github.com/gin-gonic/gin"
	"othello/pkg/play"
	"othello/pkg/rank"
	"othello/pkg/ws"
)


func Router() *gin.Engine {
	r := gin.Default()
	o := &ws.Othello{}

	// 设置html
	r.LoadHTMLGlob("assets/*")

	r.GET("/ws", o.Client)

	playGroup := r.Group("/play")
	{
		playGroup.POST("/:first/:back", func(ctx *gin.Context) {
			play.Gaming(ctx, o)
		})
		playGroup.GET("/:first/:back", play.GameOver)
	}

	rankGroup := r.Group("/rank")
	{
		rankGroup.GET("/", rank.GetRank)
		rankGroup.GET("/users", func(ctx *gin.Context) {
			rank.GetUsers(ctx, o)
		})
	}

	pageGroup := r.Group("/page")
	{
		pageGroup.GET("/show", play.Show)
		pageGroup.GET("/", func(ctx *gin.Context) {
			play.Index(ctx, o)
		})
	}


	return r
}
