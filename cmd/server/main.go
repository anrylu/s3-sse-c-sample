package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/anrylu/s3-sse-c-sample/pkg/config"
	"github.com/anrylu/s3-sse-c-sample/pkg/service"
	"go.uber.org/zap"

	"github.com/facebookgo/flagenv"
	"github.com/facebookgo/inject"
	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// process config
	qs3config := config.GetS3Config()
	flagenv.Prefix = "s3_sse_c_sample_"
	flagenv.Parse() // override from env
	flag.Parse()    // override from flag

	// create s3 service object
	qs3Service := service.QS3{}

	// dependency injection
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())
	defer rootCtxCancel()
	var g inject.Graph
	if err := g.Provide(
		&inject.Object{Value: qs3config},
		&inject.Object{Value: &qs3Service},
	); err != nil {
		log.Fatal(rootCtx, "Provide failed", zap.Error(err))
	}
	if err := g.Populate(); err != nil {
		log.Fatal(rootCtx, "Populate failed", zap.Error(err))
	}

	// init s3 service object
	qs3Service.Setup(rootCtx)

	// setup gin
	router := gin.Default()

	// upload file
	router.POST("/upload", func(c *gin.Context) {
		if err := qs3Service.Upload(c); err != nil {
			c.String(http.StatusBadRequest, "")
		}
	})

	// download file
	router.GET("/download/:filename", func(c *gin.Context) {
		if err := qs3Service.Download(c, c.Param("filename")); err != nil {
			c.String(http.StatusBadRequest, "")
		}
	})

	router.Run(":8080")
}
