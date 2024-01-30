package main

import (
	"context"
	ctxutil "github.com/tyeryan/l-protocol/context"
	logutil "github.com/tyeryan/l-protocol/log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	ctx := ctxutil.NewContext(ctxutil.WithStan("main-process"))
	log := logutil.GetLogger("lake-go")

	log.Infow(ctx, "starting service lake-go")

	router, err := injectRoutes(ctx)
	if err != nil {
		log.Fatale(ctx, "inject router failed", err)
	}

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Infow(ctx, "starting http server")
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatale(ctx, "failed to start http server", err)
		}
	}()

	go func() {
		termChan := make(chan os.Signal)
		signal.Notify(termChan, syscall.SIGUSR2)
		stack := make([]byte, 1<<20)
		for {
			<-termChan
			stackLength := runtime.Stack(stack, true)
			log.Infow(ctx, "stack trace", "dump", string(stack[:stackLength]))
		}
	}()

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	log.Infow(ctx, "stopping service", "signal", <-termChan)

	log.Infow(ctx, "stopping http server")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Errore(ctx, "failed to stop http server", err)
	}

	log.Infow(ctx, "stopped service lake-go")
}
