package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RavisMsk/xmcompanies/internal/api/components"
)

func main() {
	configPath := flag.String("config", "", "yaml config path")
	flag.Parse()

	if len(*configPath) < 1 {
		log.Fatalf("provide config path with -config")
	}

	assembly, err := components.InitializeAssembly(*configPath)
	if err != nil {
		log.Fatalf("error initializing app: %s", err)
	}

	assembly.Run()
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig

	stopped := make(chan struct{})
	go func() {
		assembly.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
		assembly.Log.Info("gracefully stopped")
	case <-time.After(30 * time.Second):
		assembly.Log.Fatal("gracefull shutdown timeout")
	}
}
