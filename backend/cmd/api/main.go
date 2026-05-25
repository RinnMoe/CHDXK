package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"jcourse/config"
	"jcourse/internal/app"
	"jcourse/internal/domain/task"
	infratask "jcourse/internal/infrastructure/task"
	"jcourse/internal/interface/web"
	"jcourse/pkg/logx"
)

func main() {
	logx.ConfigureDefault()
	ctx := context.Background()

	pflag.String("config", "", "path to config file (env: CONFIG_PATH)")
	pflag.Parse()

	if err := viper.BindPFlag("config", pflag.Lookup("config")); err != nil {
		logx.Fatal(ctx, "bind config flag", "err", err)
	}
	if err := viper.BindEnv("config", "CONFIG_PATH"); err != nil {
		logx.Fatal(ctx, "bind config env", "err", err)
	}

	configPath := viper.GetString("config")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	conf, err := config.Load(configPath)
	if err != nil {
		logx.Fatal(ctx, "failed to load config", "err", err)
	}

	container := app.NewServiceContainer(conf)
	engine := web.NewRouter(container, conf)

	taskClient := infratask.NewClient(conf.Redis)
	defer func() {
		if err := taskClient.Close(); err != nil {
			logx.Warn(ctx, "close task client", "err", err)
		}
	}()
	task.SetEnqueuer(infratask.NewEnqueuer(taskClient))

	srv := &http.Server{
		Addr:    conf.Server.Addr,
		Handler: engine,
	}

	go func() {
		logx.Info(ctx, "server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.Fatal(ctx, "listen", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logx.Info(ctx, "shutting down server")

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logx.Fatal(ctx, "server forced to shutdown", "err", err)
	}
	logx.Info(ctx, "server exited")
}
