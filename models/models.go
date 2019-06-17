package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 用户表
type UserProfile struct {
	gorm.Model
	Name       string `gorm:"size:255;not null;"`
	Email      string `gorm:"not null;unique"`
	Password   string `gorm:"not null"`
	Avatar     string `gorm:"default:'/static/images/avatar.jpg'"`
	Background string `gorm:"default:'/static/images/background.jpg'"`
	Info       string
	Birthday   time.Time

	// 外键区域
	Blogs    []Blog    `gorm:"ForeignKey:UserId"` // 用户发送的博客
	Comments []Comment `gorm:"ForeignKey:UserId"` // 用户发送的评论
}

// 用户发送的微博
type Blog struct {
	gorm.Model
	UserId     uint // 外键关联，发送微博的用户
	Content    string `gorm:"not null"`
	ThumbUpNum int64  `gorm:"default:0"` // 点赞数
	CommentNum int64  `gorm:"default:0"` // 评论数

	// 外键区域
	Comments []Comment   `gorm:"ForeignKey:BlogId"` // 微博的评论
	Photos   []BlogPhoto `gorm:"ForeignKey:BlogId"` // 微博附带的照片
}

// 微博图片
type BlogPhoto struct {
	PhotoPath string // 微博图片路径
	BlogId    uint   // 外键关联， 博客id
}

// 评论表
type Comment struct {
	gorm.Model
	BlogId   uint                      // 外键关联，被评论的博客
	UserId   uint                      // 外键关联，提交评论的用户
	Content  string `gorm:"not null"`  // 评论内容
	ParentId uint   `gorm:"default:0"` // 父评论
}

// 关注表
type Follow struct {
	gorm.Model
	UserId   uint
	FriendId uint
}

// 通知
type Notification struct {
	gorm.Model
	Title   string `gorm:"not null"` // 通知标题
	Content string `gorm:"not null"` // 通知内容
}
