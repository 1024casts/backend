package course

import (
	"github.com/1024casts/1024casts/model"
	"github.com/1024casts/1024casts/pkg/app"
	"github.com/1024casts/1024casts/pkg/errno"
	"github.com/1024casts/1024casts/service"
	"github.com/1024casts/1024casts/util"

	"github.com/gin-gonic/gin"
	"github.com/lexkong/log"
	"github.com/lexkong/log/lager"
)

// @Summary Add new course to the database
// @Description Add a new course
// @Tags course
// @Accept  json
// @Produce  json
// @Param course body course.CreateRequest true "Create a new course"
// @Success 200 {object} course.CreateResponse "{"code":0,"message":"OK","data":{"name":"test"}}"
// @Router /courses [post]
func Create(c *gin.Context) {
	log.Info("Course Create function called.", lager.Data{"X-Request-Id": util.GetReqID(c)})
	var r CreateRequest
	if err := c.Bind(&r); err != nil {
		app.Response(c, errno.ErrBind, nil)
		return
	}

	course := model.CourseModel{
		Name:         r.Name,
		Type:         r.Type,
		Keywords:     r.Keywords,
		Description:  r.Description,
		Slug:         r.Slug,
		Content:      r.Content,
		CoverKey:     r.CoverKey,
		UserId:       r.UserId,
		UpdateStatus: r.UpdateStatus,
		IsPublish:    r.IsPublish,
	}

	srv := service.NewCourseService()
	//user, err := model.GetUserById(userId)
	id, err := srv.CreateCourse(course)
	if err != nil {
		app.Response(c, errno.ErrCourseCreateFail, nil)
		log.Warn("course info", lager.Data{"id": id})
		return
	}

	resp := CreateResponse{
		Id: course.Id,
	}

	// Show the user information.
	app.Response(c, nil, resp)
}
