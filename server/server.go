package server

import (
	"github.com/SeijiOmi/posts-service/controller"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Init is initialize server
func Init() {
	r := router()
	r.Run(":8090")
}

func router() *gin.Engine {
	r := gin.Default()

	// https://godoc.org/github.com/gin-gonic/gin#RouterGroup.Use
	r.Use(cors.New(cors.Config{
		// 許可したいHTTPメソッドの一覧
		AllowMethods: []string{
			"POST",
			"GET",
			"OPTIONS",
			"PUT",
			"DELETE",
		},
		// 許可したいHTTPリクエストヘッダの一覧
		AllowHeaders: []string{
			"Access-Control-Allow-Headers",
			"X-Requested-With",
			"Origin",
			"X-Csrftoken",
			"Content-Type",
			"Accept",
		},
		// 許可したいアクセス元の一覧
		AllowOrigins: []string{
			"*",
		},
	}))

	p := r.Group("/posts")
	{
		p.GET("", controller.Index)
		p.GET("/:id", controller.Show)
		p.POST("", controller.Create)
		p.PUT("/:id", controller.Update)
		p.DELETE("/:id", controller.Delete)
	}

	u := r.Group("/user")
	{
		u.GET("/:id", controller.UserShow)
	}

	h := r.Group("/helper")
	{
		h.GET("/:id", controller.HelperShow)
		h.POST("", controller.SetHelpUser)
		h.DELETE("/:id", controller.TakeHelpUser)
	}

	d := r.Group("/done")
	{
		d.POST("", controller.DonePayment)
		d.PUT("/:id", controller.DoneAcceptance)
	}

	a := r.Group("/amount")
	{
		a.GET("/:id", controller.AmountPayment)
	}

	return r
}
