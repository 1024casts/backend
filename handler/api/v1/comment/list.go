package comment

import (
	"github.com/1024casts/1024casts/pkg/app"
	"github.com/1024casts/1024casts/pkg/errno"
	"github.com/1024casts/1024casts/service"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lexkong/log"
)

// @Summary List the comments in the database
// @Description List comments
// @Tags comment
// @Accept  json
// @Produce  json
// @Param comment body comment.ListRequest true "List courses"
// @Success 200 {object} comment.SwaggerListResponse "{"code":0,"message":"OK","data":{"totalCount":1,"userList":[{"id":0,"username":"admin","random":"user 'admin' get random string 'EnqntiSig'","password":"$2a$10$veGcArz47VGj7l9xN7g2iuT9TF21jLI1YGXarGzvARNdnt4inC9PG","createdAt":"2018-05-28 00:25:33","updatedAt":"2018-05-28 00:25:33"}]}}"
// @Router /comments [get]
func List(c *gin.Context) {
	log.Info("List function called.")
	var r ListRequest
	if err := c.Bind(&r); err != nil {
		app.Response(c, errno.ErrBind, nil)
		return
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		log.Error("get page error", err)
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		log.Error("get limit error", err)
	}

	offset := (page - 1) * limit

	srv := service.NewCommentService()

	courseMap := make(map[string]interface{})

	name := c.Query("name")
	if name != "" {
		courseMap["name"] = name
	}

	infos, count, err := srv.GetCommentList(courseMap, offset, limit)
	if err != nil {
		app.Response(c, err, nil)
		return
	}

	app.Response(c, nil, ListResponse{
		TotalCount: count,
		List:       infos,
	})
}
