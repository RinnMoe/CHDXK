package main

import (
	"fmt"
	"log"
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

	client := infratask.NewClient(conf.Redis)
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("close task client: %v", err)
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
			log.Fatalf("register site stats scheduler: %v", err)
		}
		if err := scheduler.Start(); err != nil {
			log.Fatalf("start scheduler: %v", err)
		}
	}

	server := infratask.NewServer(conf.Redis, conf.Asynq)
	mux := async.NewMux(container)

	fmt.Println("task worker starting...")
	if err := server.Start(mux); err != nil {
		log.Fatalf("start server: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("shutting down task worker...")
	scheduler.Shutdown()
	server.Shutdown()
	fmt.Println("task worker exited")
}

func mustLoadStatsLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatalf("load stats timezone: %v", err)
	}
	return loc
}
