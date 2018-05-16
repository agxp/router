package main

import (
	"context"
	"github.com/agxp/cloudflix/video-upload-svc/proto"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	_ "github.com/micro/go-plugins/registry/kubernetes"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
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

func (s *Router) PresignedURL(c *gin.Context) {
	log.Info("Recieved request for test")

	filename := c.PostForm("filename")

	log.Info("filename is: ", filename)

	res, err := vu.S3Request(context.Background(), &video_upload.Request{
		Filename: filename,
	})

	if err != nil {
		log.Fatal(err)
		c.JSON(500, err)
	}

	res.PresignedUrl = strings.Replace(res.PresignedUrl, MINIO_INTERNAL_URL, MINIO_EXTERNAL_URL, -1)
	log.Print(res.PresignedUrl)

	c.JSON(200, res.PresignedUrl)
}

func main() {

	// Create a new service. Optionally include some options here.
	//service := k8s.NewService(
	// This name must match the package name given in your protobuf definition
	//web.Name("router"),
	//)

	// 	index, err := ioutil.ReadFile("./static/index.html")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	Init will parse the command line flags.
	// 	srv.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
	// 		log.Info("Reached blahblahblahfuck")
	// 		http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader(index))
	// 	})

	// 	srv.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	// 		log.Info("reached /presignedURL")
	// 		if r.Method == "POST" {
	// 			log.Info("method == POST")
	// 			if err := r.ParseForm(); err !bin/magento setup:static-content:deploy= nil {
	// 				log.Fatal("ERROR IN PARSEFORM: %s", err)
	// 			}
	// 			log.Info("trying to get filename")
	// 			filename := r.Form.Get("filename")
	// 			log.Info("filename", filename)

	// 			// 			vu = video_upload.NewUploadClient("video_upload", client.DefaultClient)

	// 			// 			rsp, err := vu.S3Request(context.Background(), &video_upload.Request{
	// 			// 				Filename: filename,
	// 			// 			})

	// 			if err != nil {
	// 				http.Error(w, err.Error(), 500)
	// 				return
	// 			}

	// 			// 				req := client.NewRequest("video-upload", "Upload.S3Request", &video_upload.Request{
	// 			// 					Filename: r.PostFormValue("filename"),
	// 			// 				})

	// 			// 				rsp := &video_upload.Response{}

	// 			// 				if err := client.Call(context.Background(), req, rsp); err != nil {
	// 			// 					log.Fatal(err, rsp)
	// 			// 				}
	// 			w.Write([]byte("test"))
	// 			return
	// 		}
	// 		fmt.Fprint(w, `error`)

	// 	})

	//if err := service.Init(); err != nil {
	//	log.Fatal(err)
	//}

	// setup video upload service client
	//vu = video_upload.NewUploadClient("video_upload", client.DefaultClient)

	r := new(Router)
	router := gin.Default()
	router.Static("/upload", "./static/")
	router.POST("/presignedURL", r.PresignedURL)
	router.NoRoute(func(c *gin.Context) {
		c.String(404, "not found")
	})

	//service.Handle("/", router)
	router.Run()
	log.Info("Started router with minikube holy shit!!!")

	// Run the server
	//if err := service.Run(); err != nil {
	//	log.Fatal(err)
	//}
}
