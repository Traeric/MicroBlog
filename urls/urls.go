package urls

import (
	"../views"
	"github.com/labstack/echo"
)

func Urls_pattern(e *echo.Echo) {
	e.GET("/", views.Index)
	e.GET("/login", views.Login)
	e.POST("/login", views.Login)
	e.POST("/register", views.Register)
	e.GET("/home_page/:user_id", views.HomePage)
	e.GET("/following/:id", views.Following)
	e.GET("/follower/:id", views.Follower)
	e.GET("/chat", views.Chat)
	e.POST("/send_blog", views.SendBlog)
	e.POST("/add_comment", views.AddComment)
	e.DELETE("/delete_comment/:id/:blog_id", views.DeleteComment)
	e.DELETE("/delete_blog/:blog_id", views.DeleteBlog)
	e.POST("/follow/:user_id/:follower_id", views.Follow)
	e.DELETE("/unfollow/:user_id/:follower_id", views.UnFollow)
	e.GET("/settings", views.Settings)
}
