package proxyapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

// SetVideoJobService enables the /v1/video/* routes.
func (c *ProxyController) SetVideoJobService(service *VideoJobService) {
	c.videoJobs = service
}

func (c *ProxyController) registerVideoRoutes(r interfaces.Router) {
	r.POST("/v1/video/jobs", c.authConfig.CheckAuth("video_jobs"), c.CreateVideoJob)
	r.GET("/v1/video/jobs", c.authConfig.CheckAuth("video_jobs"), c.ListVideoJobs)
	r.GET("/v1/video/jobs/:id", c.authConfig.CheckAuth("video_jobs"), c.GetVideoJob)
	r.POST("/v1/video/quote", c.authConfig.CheckAuth("video_jobs"), c.VideoQuote)
}

type VideoJobListRes struct {
	Object string         `json:"object" example:"list"`
	Data   []videojob.Job `json:"data"`
}

type VideoQuoteRes struct {
	Object string       `json:"object" example:"list"`
	Data   []VideoQuote `json:"data"`
}

func (c *ProxyController) videoJobsOr503(ctx *gin.Context) *VideoJobService {
	if c.videoJobs == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": ErrVideoJobsDisabled.Error()})
		return nil
	}
	return c.videoJobs
}

func videoErrorStatus(err error) int {
	switch {
	case errors.Is(err, videojob.ErrInvalidRequest), errors.Is(err, videojob.ErrQuoteTooLong):
		return http.StatusBadRequest
	case errors.Is(err, ErrVideoJobNotFound), errors.Is(err, ErrSessionNotFound), errors.Is(err, ErrProviderNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrProviderRefused):
		return http.StatusUnprocessableEntity
	}
	return http.StatusBadGateway
}

// CreateVideoJob godoc
//
//	@Summary		Submit a video job
//	@Description	Sends the job on a session already opened for it and returns the job once the vendor accepted it. Open the session first with POST /blockchain/bids/{id}/session, for the bid and session_seconds from POST /v1/video/quote. A session takes several jobs while their quotes add up to no more than its length. Poll GET /v1/video/jobs/{id} for the video. Only prompt is required; other fields take the provider guide's defaults. A failed submit or render leaves the session open. Close it with POST /blockchain/sessions/{id}/close when done: a clip not delivered by then is not handed over.
//	@Tags			video
//	@Accept			json
//	@Produce		json
//	@Param			session_id	header		string					true	"Session opened for this job"	format(hex32)
//	@Param			request		body		videojob.CreateRequest	true	"Video job"
//	@Success		201			{object}	videojob.Job
//	@Security		BasicAuth
//	@Router			/v1/video/jobs [post]
func (c *ProxyController) CreateVideoJob(ctx *gin.Context) {
	service := c.videoJobsOr503(ctx)
	if service == nil {
		return
	}

	var head PromptHead
	if err := ctx.ShouldBindHeader(&head); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var body videojob.CreateRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestID := head.RequestID
	if requestID == "" {
		requestID = lib.GenerateRequestID()
	}
	ctx.Writer.Header().Set("X-Request-Id", requestID)
	reqCtx := lib.ContextWithRequestID(ctx, requestID)

	job, err := service.Create(reqCtx, CreateVideoJobParams{
		Request:   body,
		SessionID: head.SessionID.Hash,
	})
	if err != nil {
		c.log.Warnf("video job not created: %s", err)
		ctx.JSON(videoErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, job)
}

// ListVideoJobs godoc
//
//	@Summary	List video jobs
//	@Tags		video
//	@Produce	json
//	@Success	200	{object}	proxyapi.VideoJobListRes
//	@Security	BasicAuth
//	@Router		/v1/video/jobs [get]
func (c *ProxyController) ListVideoJobs(ctx *gin.Context) {
	service := c.videoJobsOr503(ctx)
	if service == nil {
		return
	}
	ctx.JSON(http.StatusOK, VideoJobListRes{Object: "list", Data: service.List()})
}

// GetVideoJob godoc
//
//	@Summary		Get a video job
//	@Description	The job's status as this node last read it. video_url appears once the job succeeded; it is a bearer link that expires (expires_at, when the vendor documents it).
//	@Tags			video
//	@Produce		json
//	@Param			id	path		string	true	"Video job ID"
//	@Success		200	{object}	videojob.Job
//	@Security		BasicAuth
//	@Router			/v1/video/jobs/{id} [get]
func (c *ProxyController) GetVideoJob(ctx *gin.Context) {
	service := c.videoJobsOr503(ctx)
	if service == nil {
		return
	}
	job, err := service.Get(ctx.Param("id"))
	if err != nil {
		ctx.JSON(videoErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, job)
}

// VideoQuote godoc
//
//	@Summary		Quote a video job
//	@Description	The session length (and its cost at the bid price) each provider would sell for this job. The stake to lock comes from GET /blockchain/models/{id}/session/estimate with that length.
//	@Tags			video
//	@Accept			json
//	@Produce		json
//	@Param			model_id	header		string					false	"Model ID, when the body has no model"	format(hex32)
//	@Param			request		body		videojob.CreateRequest	true	"Video job"
//	@Success		200			{object}	proxyapi.VideoQuoteRes
//	@Security		BasicAuth
//	@Router			/v1/video/quote [post]
func (c *ProxyController) VideoQuote(ctx *gin.Context) {
	service := c.videoJobsOr503(ctx)
	if service == nil {
		return
	}
	var head PromptHead
	if err := ctx.ShouldBindHeader(&head); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var body videojob.CreateRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	modelID := head.ModelID.Hash
	if body.Model != "" {
		parsed, err := lib.HexToHash(body.Model)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid model: %s", err)})
			return
		}
		modelID = parsed
	}
	if modelID == (common.Hash{}) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	quotes, err := service.Quote(ctx, modelID, &body)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, VideoQuoteRes{Object: "list", Data: quotes})
}
