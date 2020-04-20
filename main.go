package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"test.com/video/functions"
)

func processRequest(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Upload error: %s", err.Error()))
		return
	}

	contentType := file.Header["Content-Type"][0]
	if contentType != "application/octet-stream" {
		c.String(http.StatusBadRequest, "Invalid file format")
		return
	}

	buf := bytes.NewBuffer(nil)
	stream, err := file.Open()
	if errOcured := functions.ErrProcess(err, c); errOcured {
		return
	}
	io.Copy(buf, stream)
	content := buf.Bytes()

	extension := file.Filename[strings.Index(file.Filename, ".")+1:]
	config, err := functions.FindConfigByExtension(extension)
	if errOcured := functions.ErrProcess(err, c); errOcured {
		return
	}

	media := config.WithContent(content).Parse()
	media.Name = file.Filename[:strings.Index(file.Filename, ".")]
	media.Extension = extension

	c.JSON(http.StatusOK, media)
}

func main() {
	router := gin.Default()

	router.MaxMultipartMemory = 100000 << 20
	router.POST("/process", processRequest)
	router.Run(":4000")
}
