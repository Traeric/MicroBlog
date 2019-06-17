package views

import (
	"../models"
	"../settings"
	"crypto/md5"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	DB *gorm.DB
)

type ResultMsg struct {
	Flag bool   `json:"flag"`
	Msg  string `json:"msg"`
}

func init() {
	// 加载数据库
	var err error
	DB, err = gorm.Open("sqlite3", "db.sqlite3")
	if err != nil {
		panic("数据库连接失败..")
	}
	// 自动迁移模式
	DB.AutoMigrate(&models.UserProfile{}, &models.Blog{}, &models.Comment{},
		&models.Follow{}, &models.Notification{}, &models.BlogPhoto{})
}

// 登陆
func Login(c echo.Context) error {
	request := c.Request()
	if request.Method == "GET" {
		return c.Render(http.StatusOK, "login.html", nil)
	} else if request.Method == "POST" {
		// 获取登陆信息
		email := c.FormValue("email")
		password := c.FormValue("password")
		// 进行md5加密
		data := []byte(password)
		hex := md5.Sum(data)
		password = fmt.Sprintf("%x", hex)
		// 查看是否存在
		var user_profile models.UserProfile
		DB.Where(map[string]interface{}{"email": email, "password": password}).First(&user_profile)
		if user_profile.ID > 0 {
			// 账号正确
			// 将账号信息保存到session中
			var store = sessions.NewCookieStore([]byte(settings.BaseConfigDomain.SessionKey))
			session, _ := store.Get(c.Request(), "userInfo") // userInfo是储存在浏览器端的cookie的名字
			// 设置session有效期
			session.Options = &sessions.Options{
				Path:   "/",
				MaxAge: 30 * 24 * 60 * 60, // 保存一个月
			}
			session.Values["id"] = user_profile.ID
			session.Values["name"] = user_profile.Name
			session.Values["email"] = user_profile.Email
			session.Values["avatar"] = user_profile.Avatar
			session.Values["background"] = user_profile.Background
			session.Values["info"] = user_profile.Info
			session.Values["birth"] = user_profile.Birthday.Format("2006年01月02日 15:05")
			session.Values["create"] = user_profile.CreatedAt.Format("2006年01月02日 15:05")
			err := session.Save(c.Request(), c.Response()) // 保存session
			if err != nil {
				http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
				return nil
			}
			// 跳转
			return c.Redirect(http.StatusMovedPermanently, "/")
		} else {
			// 账号错误
			return c.Render(http.StatusMovedPermanently, "login.html", map[string]string{
				"err_msg": "邮箱或密码错误！",
			})
		}
	} else {
		return c.String(http.StatusOK, "NOT FOUND")
	}
}

// 注册
func Register(c echo.Context) error {
	// 获取注册参数
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	// 检查是否注册
	var user_profile models.UserProfile
	DB.Where(map[string]interface{}{"email": email}).First(&user_profile)
	if user_profile.ID > 0 {
		// 账号已经存在
		return c.Render(http.StatusMovedPermanently, "login.html", map[string]string{
			"err_msg":  "账号已经存在，可直接登陆",
			"email":    email,
			"password": password,
		})
	} else {
		// 不存在，创建
		// 将密码进行md5加密
		data := []byte(password)
		hex := md5.Sum(data)
		md5_password := fmt.Sprintf("%x", hex)

		var user_profile = &models.UserProfile{
			Name:     name,
			Email:    email,
			Password: md5_password,
		}
		DB.Create(user_profile)
		// 创建成功，登陆
		return c.Render(http.StatusMovedPermanently, "login.html", map[string]interface{}{
			"email":    email,
			"password": password,
		})
	}
}

// 首页
func Index(c echo.Context) error {
	// 读取用户信息
	store := sessions.NewCookieStore([]byte(settings.BaseConfigDomain.SessionKey))
	session, _ := store.Get(c.Request(), "userInfo")
	// 统计用户的微博数目
	var blog_num int
	DB.Model(&models.Blog{}).Where("user_id = ?", session.Values["id"]).Count(&blog_num)
	// 统计用户的关注者和正在关注的人
	var follower_num int
	var following_num int
	DB.Model(&models.Follow{}).Where("user_id = ?", session.Values["id"]).Count(&following_num)
	DB.Model(&models.Follow{}).Where("friend_id = ?", session.Values["id"]).Count(&follower_num)
	// 获取前十条微博
	var blog_data = make([]models.Blog, 10)
	DB.Order("id desc").Limit(10).Find(&blog_data)
	// 整理数据
	var blogs []map[string]interface{}
	for index, value := range blog_data {
		item := map[string]interface{}{
			"id":         value.ID,
			"content":    value.Content,
			"ThumbUpNum": value.ThumbUpNum,
			"CommentNum": value.CommentNum,
			"createAt":   value.CreatedAt.Format("2006年01月02日 15:05"),
			"index":      index,
		}
		// 查询发布微博的用户信息
		var user models.UserProfile
		DB.First(&user, value.UserId)
		item["user"] = user
		// 查询微博的照片
		DB.Model(&value).Related(&value.Photos)
		item["Photos"] = value.Photos
		// 查询评论
		DB.Model(&value).Related(&value.Comments)
		item["Comments"] = value.Comments
		// 布局评论
		// 将所有的评论从父级往下一级一级排好
		var comments []models.Comment
		DB.Find(&comments, "blog_id = ?", value.ID)
		comment_list := LoadComment(comments)
		item["comments"] = template.HTML(LoadCommentToString(comment_list))
		blogs = append(blogs, item)
	}

	return c.Render(http.StatusMovedPermanently, "index.html", map[string]interface{}{
		"id":            session.Values["id"],
		"name":          session.Values["name"],
		"email":         session.Values["email"],
		"avatar":        session.Values["avatar"],
		"background":    session.Values["background"],
		"info":          session.Values["info"],
		"blog_num":      blog_num,
		"follower_num":  follower_num,
		"following_num": following_num,
		"blogs":         blogs,
	})
}

// 个人主页
func HomePage(c echo.Context) error {
	// 读取用户信息
	store := sessions.NewCookieStore([]byte(settings.BaseConfigDomain.SessionKey))
	session, _ := store.Get(c.Request(), "userInfo")

	return c.Render(http.StatusMovedPermanently, "home_page.html", map[string]interface{}{
		"id":         session.Values["id"],
		"name":       session.Values["name"],
		"email":      session.Values["email"],
		"avatar":     session.Values["avatar"],
		"background": session.Values["background"],
		"info":       session.Values["info"],
		"birth":      session.Values["birth"],
		"create":     session.Values["create"],
	})
}

// 正在关注
func Following(c echo.Context) error {
	user_id := c.Param("id") // 获取用户id
	fmt.Println(user_id)
	// 查询用户正在关注的人的信息

	store := sessions.NewCookieStore([]byte(settings.BaseConfigDomain.SessionKey))
	session, _ := store.Get(c.Request(), "userInfo")
	return c.Render(http.StatusMovedPermanently, "following.html", map[string]interface{}{
		"id":         session.Values["id"],
		"name":       session.Values["name"],
		"email":      session.Values["email"],
		"avatar":     session.Values["avatar"],
		"background": session.Values["background"],
		"info":       session.Values["info"],
		"birth":      session.Values["birth"],
		"create":     session.Values["create"],
	})
}

// 关注者
func Follower(c echo.Context) error {
	user_id := c.Param("id") // 获取用户id
	fmt.Println(user_id)
	// 查询用户正在关注的人的信息

	store := sessions.NewCookieStore([]byte(settings.BaseConfigDomain.SessionKey))
	session, _ := store.Get(c.Request(), "userInfo")
	return c.Render(http.StatusMovedPermanently, "follower.html", map[string]interface{}{
		"id":         session.Values["id"],
		"name":       session.Values["name"],
		"email":      session.Values["email"],
		"avatar":     session.Values["avatar"],
		"background": session.Values["background"],
		"info":       session.Values["info"],
		"birth":      session.Values["birth"],
		"create":     session.Values["create"],
	})
}

// 私信
func Chat(c echo.Context) error {
	store := sessions.NewCookieStore([]byte(settings.BaseConfigDomain.SessionKey))
	session, _ := store.Get(c.Request(), "userInfo")
	return c.Render(http.StatusMovedPermanently, "chat.html", map[string]interface{}{
		"id":         session.Values["id"],
		"name":       session.Values["name"],
		"email":      session.Values["email"],
		"avatar":     session.Values["avatar"],
		"background": session.Values["background"],
		"info":       session.Values["info"],
		"birth":      session.Values["birth"],
		"create":     session.Values["create"],
	})
}

// 上传微博
func SendBlog(c echo.Context) error {
	// 读取用户信息
	store := sessions.NewCookieStore([]byte(settings.BaseConfigDomain.SessionKey))
	session, _ := store.Get(c.Request(), "userInfo")
	// 获取参数
	content := c.FormValue("content")
	// 检查内容不能为空
	if strings.TrimSpace(content) == "" {
		return c.JSON(http.StatusOK, ResultMsg{
			Flag: false,
			Msg:  "内容不能为空",
		})
	}
	// 创建一条记录
	var blog = &models.Blog{
		Content: content,
		UserId:  session.Values["id"].(uint),
	}
	DB.Create(blog)

	// 读取上传图片
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusOK, ResultMsg{
			Flag: false,
			Msg:  "图片上传错误",
		})
	}
	files := form.File["photos"]

	// 保存图片
	for _, file := range files {
		// 上传的文件
		src, _ := file.Open()
		defer src.Close()
		path := fmt.Sprintf(`%s/%s/%d/%d/`, settings.BaseDir, "blog_photo", session.Values["id"], blog.ID)
		_, err := os.Stat(path)
		// 目录不存在
		if os.IsNotExist(err) {
			// 创建目录
			_ = os.MkdirAll(path, os.ModePerm)
		}
		// 存储图片的地址
		dst, _ := os.Create(path + file.Filename)
		defer dst.Close()
		// Copy
		_, err = io.Copy(dst, src)
		if err == nil {
			// 在数据库创建记录
			var blogPhoto = &models.BlogPhoto{
				PhotoPath: fmt.Sprintf(`/static/userInfo/blog_photo/%d/%d/%s`, session.Values["id"], blog.ID, file.Filename),
				BlogId:    blog.ID,
			}
			DB.Create(blogPhoto)
		}
	}

	return c.JSON(http.StatusOK, ResultMsg{
		Flag: true,
		Msg:  "发表成功",
	})
}

// 添加评论
func AddComment(c echo.Context) error {
	// 获取用户
	store := sessions.NewCookieStore([]byte(settings.BaseConfigDomain.SessionKey))
	session, _ := store.Get(c.Request(), "userInfo")
	// 获取参数
	comment := c.FormValue("comment")
	blog_id, _ := strconv.Atoi(c.FormValue("blog_id"))
	comment_id, _ := strconv.Atoi(c.FormValue("comment_id"))
	if comment := strings.TrimSpace(comment); comment == "" {
		return c.JSON(http.StatusOK, &ResultMsg{
			Flag: false,
			Msg:  "评论内容不能为空",
		})
	}

	// 保存评论信息
	DB.Create(&models.Comment{
		BlogId:   uint(blog_id),
		UserId:   session.Values["id"].(uint),
		Content:  comment,
		ParentId: uint(comment_id),
	})

	// 增加一条评论数
	var blog models.Blog
	DB.First(&blog, "id = ?", uint(blog_id))
	// 评论数加一
	comment_num := blog.CommentNum + 1
	DB.Model(&blog).Where("id = ?", uint(blog_id)).Update("comment_num", comment_num)

	return c.JSON(http.StatusOK, &ResultMsg{
		Flag: true,
		Msg:  "评论成功",
	})
}

// 删除评论
func DeleteComment(c echo.Context) error {
	// 获取评论id
	comment_id := c.Param("id")
	blog_id, _ := strconv.Atoi(c.Param("blog_id"))
	var comment models.Comment
	DB.First(&comment, "id = ?", comment_id)
	// 删除
	DB.Delete(&comment)

	// 减少一条评论数
	var blog models.Blog
	DB.First(&blog, "id = ?", uint(blog_id))
	// 评论数减一
	comment_num := blog.CommentNum - 1
	DB.Model(&blog).Where("id = ?", uint(blog_id)).Update("comment_num", comment_num)

	return c.JSON(http.StatusOK, &ResultMsg{
		Flag: true,
		Msg:  "评论删除成功!",
	})
}
