package user

import (
	"html/template"
	"net/http"

	"github.com/1024casts/1024casts/pkg/constvar"

	"github.com/1024casts/1024casts/pkg/pagination"

	"strconv"

	"github.com/1024casts/1024casts/service"
	"github.com/1024casts/1024casts/util"
	"github.com/gin-gonic/gin"
	"github.com/lexkong/log"
)

func TopicList(c *gin.Context) {
	userId := util.GetUserId(c)
	srv := service.NewUserService()
	user, err := srv.GetUserById(userId)
	if err != nil {
		log.Warnf("[user] get user info err, %v", err)
	}

	userName := c.Param("username")
	userInfo, err := srv.GetUserByUsername(userName)
	if err != nil {
		c.HTML(http.StatusNotFound, "error/404", gin.H{
			"title": "404错误",
			"ctx":   c,
		})
		return
	}
	topicSrv := service.NewTopicService()

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		log.Error("get page error", err)
	}
	limit := constvar.DefaultLimit
	offset := (page - 1) * limit
	topicMap := make(map[string]interface{})
	topicMap["user_id"] = userInfo.Id
	topics, count, err := topicSrv.GetTopicList(topicMap, offset, limit)
	if err != nil {
		log.Warnf("[topic] get topic list err: %v", err)
	}

	pagination := pagination.NewPagination(c.Request, count, limit)

	c.HTML(http.StatusOK, "user/topicList", gin.H{
		"title":    "我发布的话题",
		"user_id":  userId,
		"user":     user,
		"userInfo": userInfo,
		"ctx":      c,
		"topics":   topics,
		"pages":    template.HTML(pagination.Pages()),
	})
}
