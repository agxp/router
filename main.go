package main

import (
	"github.com/gin-gonic/gin"
	"github.com/micro/go-web"
	k8s "github.com/micro/kubernetes/go/web"
	log "github.com/sirupsen/logrus"
	"github.com/agxp/cloudflix/video-upload-svc/proto"
	"context"
	"github.com/micro/go-micro/client"
)

type Router struct{}

var (
	vu video_upload.UploadClient
)

func (s *Router) PresignedURL(c *gin.Context) {
	log.Info("Recieved request for test")

	filename := c.PostForm("filename")

	log.Info("filename is: ", filename)

	res, err := vu.S3Request(context.Background(), &video_upload.Request{
		Filename: filename,
	})

	if err != nil {
		c.JSON(500, err)
	}

	c.JSON(200, res.PresignedUrl)
}

func main() {
	// Create a new service. Optionally include some options here.
	srv := k8s.NewService(
		// This name must match the package name given in your protobuf definition
		web.Name("cloudflix.api.router"),
	)

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
	// 			if err := r.ParseForm(); err != nil {
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

	if err := srv.Init(); err != nil {
		log.Fatal(err)
	}

	// setup video upload service client
	vu = video_upload.NewUploadClient("cloudflix.api.video_upload", client.DefaultClient)



	r := new(Router)
	router := gin.Default()
	router.Static("/upload", "./static/")
	router.POST("/presignedURL", r.PresignedURL)
	router.NoRoute(func(c *gin.Context) {
		c.String(404, "not found")
	})

	srv.Handle("/", router)
	log.Info("Started router with minikube holy shit!!!")

	// Run the server
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
