package main

import (
	"github.com/agxp/cloudflix/video-upload-svc/proto"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"context"
	"net/http"
)

var MINIO_EXTERNAL_URL = os.Getenv("MINIO_EXTERNAL_URL")
var MINIO_INTERNAL_URL = "http://minio:9000"

type Router struct{}

var (
	vu video_upload.UploadClient
)

func init() {
	cmd.Init()

	vu = video_upload.NewUploadClient("video_upload", client.DefaultClient)

}
type FilenamePOST struct {
	Filename string `form:"filename" json:"filename" binding:"required"`
}

type UploadVideoPOST struct {
	Filename string `form:"filename" json:"filename" binding:"required"`
	Title string `form:"title" json:"title" binding:"required"`
	Description string `form:"description" json:"description" binding:"required"`
}

type UploadFinishPOST struct {
	Id string `form:"id" json:"id" binding:"required"`
}

func (s *Router) PresignedURL(c *gin.Context) {
	log.Info("Recieved request for PresignedURL")

	var form FilenamePOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("filename is: ", form.Filename)

		res, err := vu.S3Request(context.Background(), &video_upload.Request{
			Filename: form.Filename,
		})

		if err != nil {
			log.Fatal(err)
			c.JSON(500, err)
		}

		res.PresignedUrl = strings.Replace(res.PresignedUrl, MINIO_INTERNAL_URL, MINIO_EXTERNAL_URL, -1)
		log.Print(res.PresignedUrl)

		c.JSON(200, res)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func (s *Router) UploadFile(c *gin.Context) {
	log.Info("Recieved request for UploadFile")

	var form UploadVideoPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("filename is: ", form.Filename)

		res, err := vu.UploadFile(context.Background(), &video_upload.UploadRequest{
			Filename: form.Filename,
			Title: form.Title,
			Description: form.Description,
		})

		if err != nil {
			log.Fatal(err)
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
	log.Info("Received request for UploadFinish")

	var form UploadFinishPOST

	if err := c.ShouldBind(&form); err == nil {
		log.Info("id is: ", form.Id)

		res, err := vu.UploadFinish(context.Background(), &video_upload.UploadFinishRequest{
			Id: form.Id,
		})

		if err != nil {
			log.Fatal(err)
			c.JSON(500, err)
		}

		c.JSON(200, res)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

func main() {

	r := new(Router)
	router := gin.Default()
	router.Static("/upload", "./static/")
	router.POST("/presignedURL", r.PresignedURL)
	router.POST("/uploadFile", r.UploadFile)
	router.POST("/uploadFinish", r.UploadFinish)
	router.NoRoute(func(c *gin.Context) {
		c.String(404, "not found")
	})

	router.Run()
	log.Info("Started router")
}
