package main

import (
	"fmt"
	"image"
	"net/http"
	"os"

	"github.com/Timur1414/asvt/pkg/logger"
	"github.com/hybridgroup/mjpeg"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

var stream *mjpeg.Stream

func main() {
	DEBUG := os.Getenv("DEBUG") == "true"
	err := logger.InitLogger(DEBUG)
	if err != nil {
		fmt.Println("Error initializing logger: ", err)
		return
	}
	defer func() {
		err = logger.Close()
		if err != nil {
			fmt.Println("Error closing logger: ", err)
		}
	}()
	log := logger.GetLogger()
	err = logger.InitAccessLogger()
	if err != nil {
		log.Error("Error initializing access logger", zap.Error(err))
		return
	}
	defer func() {
		err = logger.AccessClose()
		if err != nil {
			fmt.Println("Error closing access logger: ", err)
		}
	}()

	webcam, err := gocv.VideoCaptureFile(os.Getenv("CAMERA_STREAM_URL"))
	if err != nil {
		log.Fatal("Ошибка подключения к потоку", zap.Error(err))
		return
	}
	defer webcam.Close()

	net := gocv.ReadNet(os.Getenv("YOLO_MODEL_FILE"), "")
	if net.Empty() {
		log.Fatal("Ошибка загрузки модели")
	}
	defer net.Close()

	stream = mjpeg.NewStream()
	go http.Handle("/", stream)
	go http.ListenAndServe("0.0.0.0:8080", nil)

	img := gocv.NewMat()
	defer img.Close()

	log.Info("Сервер запущен на порту 8080")

	for {
		if ok := webcam.Read(&img); !ok || img.Empty() {
			continue
		}

		blob := gocv.BlobFromImage(img, 1/255.0, image.Pt(640, 640), gocv.NewScalar(0, 0, 0, 0), true, false)
		net.SetInput(blob, "")
		prob := net.Forward("")

		blob.Close()
		prob.Close()

		// Конвертация в JPEG и отправка клиенту
		buf, _ := gocv.IMEncode(gocv.JPEGFileExt, img)
		stream.UpdateJPEG(buf.GetBytes())
		buf.Close()
	}
}
