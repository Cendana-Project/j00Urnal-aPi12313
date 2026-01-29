package term

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/api-monolith-template/internal/constant"
	"github.com/api-monolith-template/internal/mapper"
	"github.com/api-monolith-template/internal/model/request"
	"github.com/api-monolith-template/internal/model/response"
	"github.com/api-monolith-template/internal/service/term"
	"github.com/api-monolith-template/internal/util"
)

type Controller struct {
	svc *term.Service
}

func NewController(svc *term.Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) GetCurrent(ctx *gin.Context) {
	t, err := c.svc.GetActive(ctx.Request.Context())
	if err != nil {
		util.HandleError(ctx, err)
		return
	}
	if t == nil {
		util.HandleError(ctx, constant.ErrRecordNotFound)
		return
	}

	res := response.NewResponseOK()
	res.Data = mapper.ToTermResponse(t)
	util.HandleResponse(ctx, res, nil)
}

func (c *Controller) Create(ctx *gin.Context) {
	var req request.CreateTermRequest
	if err := util.BindAndValidate(ctx, &req); err != nil {
		util.HandleError(ctx, err)
		return
	}

	t, err := c.svc.Create(ctx.Request.Context(), req.Content)
	if err != nil {
		util.HandleError(ctx, err)
		return
	}

	res := response.NewResponseOK()
	res.StatusCode = http.StatusCreated
	res.Data = mapper.ToTermResponse(t)
	util.HandleResponse(ctx, res, nil)
}
