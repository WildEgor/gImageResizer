package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	server "github.com/WildEgor/gImageResizer/internal"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

var srv *fiber.App

// @title Fiber Example API
// @version 1.0
// @description This is a sample swagger for Fiber
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
func main() {
	Start()
	Shutdown()
}

func Start() {
	srv, _ = server.NewServer()
	go func() {
		if err := srv.Listen(fmt.Sprintf(":%v", "8888")); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func Shutdown() {
	// block main thread - wait for shutdown signal
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Println()
		log.Println(sig)
		done <- true
	}()

	log.Println("[Main] Awaiting signal")
	<-done
	log.Println("[Main] Stopping consumer")
}
