package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	wg.Add(1)
	go startServer(ctx, &wg)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	cancel()

	wg.Wait()

	log.Println("graceful shutdown complete")
}

func startServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	app := gin.Default()
	app.GET("/health", func(ctx *gin.Context) {
		time.Sleep(5 * time.Second)
		ctx.String(http.StatusOK, "OK")
		return
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: app,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
		log.Println("stopping server")
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down gracefully")
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := server.Shutdown(ctxShutdown); err != nil {
			log.Fatal(err)
		}
	}
}
