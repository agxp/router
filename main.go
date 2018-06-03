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
	"github.com/agxp/cloudflix/video-search-svc/proto"
	"github.com/agxp/cloudflix/comments-svc/proto"
	"github.com/gin-contrib/cors"
)

var MINIO_EXTERNAL_URL = os.Getenv("MINIO_EXTERNAL_URL")
var MINIO_INTERNAL_URL = "http://minio:9000"

type Router struct{}

var (
	vu     video_upload.UploadClient
	vh     video_host.HostClient
	tr     trending.TrendingClient
	vs     video_search.SearchClient
	cm     comments.CommentsClient
	tracer *opentracing.Tracer
)

func init() {
	cmd.Init()

	vu = video_upload.NewUploadClient("video_upload", client.DefaultClient)
	vh = video_host.NewHostClient("video_host", client.DefaultClient)
	tr = trending.NewTrendingClient("trending", client.DefaultClient)
	vs = video_search.NewSearchClient("video_search", client.DefaultClient)
	cm = comments.NewCommentsClient("comments", client.DefaultClient)
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

type SearchPOST struct {
	Query string `form:"q" json:"q" binding:"required"`
	Page  uint64 `form:"page" json:"page"`
}

type GetCommentsFromVideoIdPOST struct {
	VideoId string `form:"video_id" json:"video_id" binding:"required"`
}

type GetCommentPOST struct {
	Id string `form:"id" json:"id" binding:"required"`
}

type WriteCommentPOST struct {
	VideoId string `form:"video_id" json:"video_id" binding:"required"`
	UserId  string `form:"user_id" json:"user_id" binding:"required"`
	Content string `form:"content" json:"content" binding:"required"`
}

func (s *Router) GetTrending(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.TODO(), "GetTrending_Route")
	defer sp.Finish()

	log.Info("Recieved request for GetTrending")

	res, err := tr.GetTrending(context.TODO(), &trending.Request{})
	if err != nil {
		log.Error(err)
		c.JSON(500, err)
	}

	log.Print(res.Data)

	c.JSON(200, res)
}

func (s *Router) UploadFile(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.TODO(), "UploadFile_Route")
	defer sp.Finish()

	log.Info("Recieved request for UploadFile")

	var form UploadVideoPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("filename is: ", form.Filename)

		res, err := vu.UploadFile(context.TODO(), &video_upload.UploadRequest{
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
	sp, _ := opentracing.StartSpanFromContext(context.TODO(), "UploadFinish_Route")
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
	sp, _ := opentracing.StartSpanFromContext(context.TODO(), "GetVideoInfo_Route")
	defer sp.Finish()

	log.Info("Received request for GetVideoInfo")

	var form GetVideoInfoPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("id is: ", form.Id)

		res, err := vh.GetVideoInfo(context.TODO(), &video_host.GetVideoInfoRequest{
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
	sp, _ := opentracing.StartSpanFromContext(context.TODO(), "GetVideo_Route")
	defer sp.Finish()

	log.Info("Received request for GetVideo")

	var form GetVideoPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("id is: ", form.Id)
		log.Info("resolution is: ", form.Resolution)

		res, err := vh.GetVideo(context.TODO(), &video_host.GetVideoRequest{
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
	sp, _ := opentracing.StartSpanFromContext(context.TODO(), "Prune_Route")
	defer sp.Finish()

	log.Info("Received request for Prune")

	res, err := tr.Prune(context.TODO(), &trending.PruneRequest{})

	if err != nil {
		log.Error(err)
		c.JSON(500, err)
	}
	c.JSON(200, res)
}

func (s *Router) Search(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.TODO(), "Search_Route")
	defer sp.Finish()

	log.Info("Received request for Search")

	var form SearchPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("Query is: ", form.Query)
		log.Info("Page is: ", form.Page)

		res, err := vs.Search(context.TODO(), &video_search.Request{
			Query: form.Query,
			Page:  form.Page,
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

func (s *Router) GetCommentsFromVideoId(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.TODO(), "GetCommentsFromVideoId_Route")
	defer sp.Finish()

	log.Info("Received request for GetCommentsFromVideoId")

	var form GetCommentsFromVideoIdPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("VideoId is: ", form.VideoId)

		res, err := cm.GetAllForVideoId(context.TODO(), &comments.Request{
			VideoId: form.VideoId,
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

func (s *Router) GetComment(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.TODO(), "GetComment_Route")
	defer sp.Finish()

	log.Info("Received request for GetComment")

	var form GetCommentPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("Comment Id is: ", form.Id)

		res, err := cm.GetSingle(context.TODO(), &comments.SingleRequest{
			CommentId: form.Id,
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

func (s *Router) WriteComment(c *gin.Context) {
	sp, _ := opentracing.StartSpanFromContext(context.TODO(), "WriteComment_Route")
	defer sp.Finish()

	log.Info("Received request for WriteComment")

	var form WriteCommentPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("VideoId is: ", form.VideoId)
		log.Info("UserId is: ", form.UserId)
		log.Info("Content is: ", form.Content)

		res, err := cm.Write(context.TODO(), &comments.WriteRequest{
			VideoId: form.VideoId,
			User:    form.UserId,
			Content: form.Content,
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
	router.Use(cors.Default())

	router.Static("/upload", "./static/")
	router.POST("/trending", r.GetTrending)
	router.POST("/uploadFile", r.UploadFile)
	router.POST("/uploadFinish", r.UploadFinish)
	router.POST("/v", r.GetVideo)
	router.POST("/videoInfo", r.GetVideoInfo)
	router.POST("/prune", r.Prune)
	router.POST("/search", r.Search)
	router.POST("/comments", r.GetCommentsFromVideoId)
	router.POST("/comment", r.GetComment)
	router.POST("/writeComment", r.WriteComment)

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
