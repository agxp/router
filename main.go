package main

import (
	"context"
	"github.com/agxp/cloudflix/video-hosting-svc/proto"
	"github.com/agxp/cloudflix/video-upload-svc/proto"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"net/http"
	"os"
	"strings"
	"github.com/agxp/cloudflix/trending-svc/proto"
)

var MINIO_EXTERNAL_URL = os.Getenv("MINIO_EXTERNAL_URL")
var MINIO_INTERNAL_URL = "http://minio:9000"

type Router struct{}

var (
	vu     video_upload.UploadClient
	vh     video_host.HostClient
	tr     trending.TrendingClient
	tracer *opentracing.Tracer
)

func init() {
	cmd.Init()

	vu = video_upload.NewUploadClient("video_upload", client.DefaultClient)
	vh = video_host.NewHostClient("video_host", client.DefaultClient)
	tr = trending.NewTrendingClient("trending", client.DefaultClient)
}

type FilenamePOST struct {
	Filename string `form:"filename" json:"filename" binding:"required"`
}

type UploadVideoPOST struct {
	Filename    string `form:"filename" json:"filename" binding:"required"`
	Title       string `form:"title" json:"title" binding:"required"`
	Description string `form:"description" json:"description" binding:"required"`
}

type UploadFinishPOST struct {
	Id string `form:"id" json:"id" binding:"required"`
}

type GetVideoInfoPOST struct {
	Id string `form:"id" json:"id" binding:"required"`
}

type GetVideoPOST struct {
	Id         string `form:"id" json:"id" binding:"required"`
	Resolution string `form:"resolution" json:"resolution" binding:"required"`
}

func (s *Router) GetTrending(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.Background(), "GetTrending_Route")
	defer sp.Finish()

	log.Info("Recieved request for GetTrending")

		res, err := tr.GetTrending(context.Background(), &trending.Request{})
		if err != nil {
			log.Error(err)
			c.JSON(500, err)
		}

		log.Print(res.Data)

		c.JSON(200, res)
}

func (s *Router) UploadFile(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.Background(), "UploadFile_Route")
	defer sp.Finish()

	log.Info("Recieved request for UploadFile")

	var form UploadVideoPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("filename is: ", form.Filename)

		res, err := vu.UploadFile(context.Background(), &video_upload.UploadRequest{
			Filename:    form.Filename,
			Title:       form.Title,
			Description: form.Description,
		})

		if err != nil {
			log.Error(err)
			c.JSON(500, err)
		}

		res.PresignedUrl = strings.Replace(res.PresignedUrl, MINIO_INTERNAL_URL, MINIO_EXTERNAL_URL, -1)
		log.Print(res.PresignedUrl)
		log.Print(res.Id)

		c.JSON(200, res)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func (s *Router) UploadFinish(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.Background(), "UploadFinish_Route")
	defer sp.Finish()

	log.Info("Received request for UploadFinish")

	var form UploadFinishPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("id is: ", form.Id)

		res, err := vu.UploadFinish(context.Background(), &video_upload.UploadFinishRequest{
			Id: form.Id,
		})

		if err != nil {
			log.Error(err)
			c.JSON(500, err)
		}

		c.JSON(200, res)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

func (s *Router) GetVideoInfo(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.Background(), "GetVideoInfo_Route")
	defer sp.Finish()

	log.Info("Received request for GetVideoInfo")

	var form GetVideoInfoPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("id is: ", form.Id)

		res, err := vh.GetVideoInfo(context.Background(), &video_host.GetVideoInfoRequest{
			Id: form.Id,
		})

		if err != nil {
			log.Error(err)
			c.JSON(500, err)
		}

		c.JSON(200, res)
	} else {
		c.JSON(http.StatusBadRequest, "Missing parameters")
	}

}

func (s *Router) GetVideo(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.Background(), "GetVideo_Route")
	defer sp.Finish()

	log.Info("Received request for GetVideo")

	var form GetVideoPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("id is: ", form.Id)
		log.Info("resolution is: ", form.Resolution)

		res, err := vh.GetVideo(context.Background(), &video_host.GetVideoRequest{
			Id:         form.Id,
			Resolution: form.Resolution,
		})

		if err != nil {
			log.Error(err)
			c.JSON(500, err)
		}

		c.JSON(200, res)
	} else {
		c.JSON(http.StatusBadRequest, "Missing parameters")
	}

}

func (s *Router) Prune(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.Background(), "Prune_Route")
	defer sp.Finish()

	log.Info("Received request for Prune")

		res, err := tr.Prune(context.Background(), &trending.PruneRequest{})

		if err != nil {
			log.Error(err)
			c.JSON(500, err)
		}
		c.JSON(200, res)
}

func main() {

	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		// parsing errors might happen here, such as when we get a string where we expect a number
		log.Printf("Could not parse Jaeger env vars: %s", err.Error())
		return
	}

	t, closer, err := cfg.NewTracer()
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	tracer = &t
	opentracing.SetGlobalTracer(t)
	defer closer.Close()

	(*tracer).StartSpan("init_tracing").Finish()

	r := new(Router)
	router := gin.Default()
	router.Static("/upload", "./static/")
	router.POST("/trending", r.GetTrending)
	router.POST("/uploadFile", r.UploadFile)
	router.POST("/uploadFinish", r.UploadFinish)
	router.POST("/v", r.GetVideo)
	router.POST("/videoInfo", r.GetVideoInfo)
	router.POST("/prune", r.Prune)
	router.NoRoute(func(c *gin.Context) {
		if c.Request.URL.EscapedPath() == "/" {
			c.String(200, "CloudFlix, under construction")
		} else {
			c.String(404, "not found")
		}
	})

	router.Run()
	log.Info("Started router")
}
