package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"jcourse/config"
	"jcourse/internal/app"
	"jcourse/internal/interface/web"
)

func main() {
	pflag.String("config", "", "path to config file (env: CONFIG_PATH)")
	pflag.Parse()

	if err := viper.BindPFlag("config", pflag.Lookup("config")); err != nil {
		log.Fatalf("bind config flag: %v", err)
	}
	if err := viper.BindEnv("config", "CONFIG_PATH"); err != nil {
		log.Fatalf("bind config env: %v", err)
	}

	configPath := viper.GetString("config")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	conf, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	container := app.NewServiceContainer(conf)
	engine := web.NewRouter(container, conf)

	srv := &http.Server{
		Addr:    conf.Server.Addr,
		Handler: engine,
	}

	go func() {
		fmt.Printf("server starting on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}
	fmt.Println("server exited")
}
