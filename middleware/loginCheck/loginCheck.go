package loginCheck

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"net/http"
	"../../settings"
	"strings"
)

func LoginCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 设置不拦截的url
			whiteUrl := []string{"/static", "/register", "/login"};
			// 检测用户是否登陆
			path := fmt.Sprintf("%s", c.Request().URL)
			// 规定不拦截的url
			for _, v := range whiteUrl {
				if strings.HasPrefix(path, v) {
					return next(c)
				}
			}
			// 获取session里面的用户信息
			store := sessions.NewCookieStore([]byte(settings.BaseConfigDomain.SessionKey))
			session, _ := store.Get(c.Request(), "userInfo")
			identify := session.Values["id"]
			if identify == nil {
				// 没有登陆
				return c.Redirect(http.StatusMovedPermanently, "/login")
			}
			// 登陆了
			return next(c)
		}
	}
}
