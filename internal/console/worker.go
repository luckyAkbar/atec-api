package console

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/luckyAkbar/atec-api/internal/config"
	"github.com/luckyAkbar/atec-api/internal/db"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/luckyAkbar/atec-api/internal/worker"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"

	"github.com/sweet-go/stdlib/mail"
	workerPkg "github.com/sweet-go/stdlib/worker"
)

var workerCMD = &cobra.Command{
	Use:  "worker",
	Long: "Start the worker server",
	Run:  workerFn,
}

func init() {
	RootCMD.AddCommand(workerCMD)
}

func workerFn(_ *cobra.Command, _ []string) {
	db.InitializePostgresConn()

	sibClient := mail.NewSendInBlueClient(config.SendInBlueSender(), config.SendinblueAPIKey(), config.SendInBlueIsActivated())
	mailgunClient := mail.NewMailgunClient(mail.MailgunConfig{
		Domain:            config.MailgunDomain(),
		PrivateKey:        config.MailgunPrivateAPIKey(),
		IsActivated:       config.MailgunIsActivated(),
		ServerSenderEmail: config.MailgunSenderEmail(),
	})

	emailRepo := repository.NewEmailRepository(db.PostgresDB)
	mailUtil := mail.NewUtility(sibClient, mailgunClient)

	server, err := worker.NewServer(config.WorkerBrokerHost(), worker.ServerConfig{
		AsynqConfig: asynq.Config{
			Concurrency:         config.WorkerConcurency(),
			Queues:              workerPkg.DefaultQueue,
			Logger:              logrus.WithField("source", "ATEC API worker server"),
			HealthCheckFunc:     workerPkg.DefaultHealtCheckFn,
			HealthCheckInterval: 5 * time.Minute,
			IsFailure:           workerPkg.DefaultIsFailureCheckerFn,
			StrictPriority:      true,
			RetryDelayFunc:      workerPkg.DefaultRetryDelayFn,
		},
		SchedulerOpts: &asynq.SchedulerOpts{
			LogLevel: config.WorkerLogLevel(),
			Logger:   logrus.New(),
			Location: time.UTC,
		},
		MailUtil: mailUtil,
		MailRepo: emailRepo,
		Limiter:  rate.NewLimiter(rate.Limit(config.WorkerLimiterLimit()), config.WorkerLimiterBurst()),
	})

	if err != nil {
		logrus.WithError(err).Fatal("failed to start worker server")
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	errch := make(chan error)

	server.Start(worker.Mux(), errch)
	select {
	case sig := <-sigCh:
		logrus.Infof("receiving signal to stop worker server from console: %s. Gracefully shutting down worker", sig.String())
		server.Stop()
		os.Exit(0)
	case err := <-errch:
		logrus.WithError(err).Error("receiving quit signal from worker server. Gracefully shutting down worker")
		server.Stop()
		os.Exit(1)
	}
}
