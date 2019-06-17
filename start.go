package main

import (
	"./middleware/loginCheck"
	"./urls"
	_ "./views"
	"github.com/labstack/echo"
	"html/template"
	"io"
)

// 加载模板
type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	t := &Template{
		templates: template.Must(template.ParseGlob("template/*.html")),
	}
	e := echo.New()
	// 配置模板
	e.Renderer = t
	// 配置静态文件
	e.Static("/static", "static")
	// 配置默认icon
	e.File("/favicon.ico", "static/images/favicon.ico")
	// 匹配url
	urls.Urls_pattern(e)
	// 登陆检测
	e.Pre(loginCheck.LoginCheck())
	// 启动
	e.Logger.Fatal(e.Start(":8000"))
}
