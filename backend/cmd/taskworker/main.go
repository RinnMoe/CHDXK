package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"jcourse/config"
	"jcourse/internal/app"
	domainstat "jcourse/internal/domain/stat"
	domaintask "jcourse/internal/domain/task"
	infratask "jcourse/internal/infrastructure/task"
	"jcourse/internal/interface/async"
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

	client := infratask.NewClient(conf.Redis)
	defer func() {
		if err := client.Close(); err != nil {
			logx.Warn(ctx, "close task client", "err", err)
		}
	}()
	domaintask.SetEnqueuer(infratask.NewEnqueuer(client))

	statsLoc := mustLoadStatsLocation()
	scheduler := infratask.NewScheduler(conf.Redis, statsLoc)
	defer scheduler.Shutdown()
	if conf.Stats.SchedulerEnabled && conf.Stats.DailyCron != "" {
		if _, err := infratask.RegisterScheduledTask(
			scheduler,
			conf.Stats.DailyCron,
			domainstat.NewCollectDailySiteStatsTask(""),
		); err != nil {
			logx.Fatal(ctx, "register site stats scheduler", "err", err)
		}
		if err := scheduler.Start(); err != nil {
			logx.Fatal(ctx, "start scheduler", "err", err)
		}
	}

	server := infratask.NewServer(conf.Redis, conf.Asynq)
	mux := async.NewMux(container)

	logx.Info(ctx, "task worker starting")
	if err := server.Start(mux); err != nil {
		logx.Fatal(ctx, "start server", "err", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logx.Info(ctx, "shutting down task worker")
	scheduler.Shutdown()
	server.Shutdown()
	logx.Info(ctx, "task worker exited")
}

func mustLoadStatsLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logx.Fatal(context.Background(), "load stats timezone", "err", err)
	}
	return loc
}
