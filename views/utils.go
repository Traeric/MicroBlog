package views

import (
	"../models"
	"fmt"
	"sort"
)

// 加载评论
func LoadComment(comment []models.Comment) []map[string]interface{} {
	var comment_list []map[string]interface{}
	comment_map := make(map[uint]interface{})
	// 获取map所有的key
	var keys []int

	// 将struct评论转换成map
	for _, item := range comment {
		var item_map = map[string]interface{}{
			"id":       item.ID,
			"createAt": item.CreatedAt.Format("2006年01月02日 15:05"),
			"content":  item.Content,
			"parentId": item.ParentId,
			"blogId":   item.BlogId,
		}
		// 加载用户信息
		var user models.UserProfile
		DB.First(&user, item.UserId)
		item_map["user"] = user
		comment_map[item.ID] = item_map
		keys = append(keys, int(item.ID))
	}

	// 对key进行排序
	sort.Ints(keys)
	for _, key := range keys {
		item := comment_map[uint(key)]
		// 如果当前评论没有父id，就把它放入根目录
		item_map := item.(map[string]interface{})
		val := item_map["parentId"]
		if (val).(uint) == 0 {
			comment_list = append(comment_list, item_map)
		} else {
			// 如果有父id，就把该条评论放到父评论的child数组里面
			var comment_map_item = comment_map[val.(uint)].(map[string]interface{})
			children, _ := comment_map_item["children"].([]map[string]interface{})
			// 保存父评论信息
			item_map["parent_name"] = comment_map_item["user"].(models.UserProfile).Name
			item_map["parent_id"] = comment_map_item["user"].(models.UserProfile).ID
			children = append(children, item_map)
			comment_map_item["children"] = children
		}
	}

	return comment_list
}

// 将评论加载成字符串
func LoadCommentToString(comment_list []map[string]interface{}) string {
	var resultStr string
	for _, item := range comment_list {
		if item["parentId"].(uint) != 0 {
			resultStr += fmt.Sprintf(`
				<div class="comment-item" style="margin-left: 100px;">
					<div class="img"><img src="%s" alt="NO IMG"></div>
					<div class="right">
						<div class="user-info clear-float">
							<div class="nick">%s 回复 %s</div>
							<div class="date">%s</div>
						</div>
						<div class="content">%s</div>
					</div>
					<div class="operation">
						<button class="layui-btn layui-btn-sm layui-btn-primary" onclick="showComment(%d, %d)">回复</button>
						<button class="layui-btn layui-btn-sm layui-btn-danger" onclick="deleteComment(%d, %d, this)">删除</button>
					</div>
				</div>
			`, item["user"].(models.UserProfile).Avatar,
				item["user"].(models.UserProfile).Name, item["parent_name"],
				item["createAt"], item["content"], item["id"], item["blogId"], item["id"], item["blogId"])
		} else {
			resultStr += fmt.Sprintf(`
				<div class="comment-item">
					<div class="img"><img src="%s" alt="NO IMG"></div>
					<div class="right">
						<div class="user-info clear-float">
							<div class="nick">%s</div>
							<div class="date">%s</div>
						</div>
						<div class="content">%s</div>
					</div>
					<div class="operation">
						<button class="layui-btn layui-btn-sm layui-btn-primary" onclick="showComment(%d, %d)">回复</button>
						<button class="layui-btn layui-btn-sm layui-btn-danger" onclick="deleteComment(%d, %d, this)">删除</button>
					</div>
				</div>
			`, item["user"].(models.UserProfile).Avatar,
				item["user"].(models.UserProfile).Name, item["createAt"],
				item["content"], item["id"], item["blogId"], item["id"], item["blogId"])
		}
		// 查看是否有children
		if children, ok := item["children"]; ok {
			resultStr += LoadCommentToString(children.([]map[string]interface{}))
		}
	}
	return resultStr
}
