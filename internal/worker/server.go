package worker

import (
	"github.com/hibiken/asynq"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/sweet-go/stdlib/mail"
	workerPkg "github.com/sweet-go/stdlib/worker"
	"golang.org/x/time/rate"
)

var mux = asynq.NewServeMux()

func registerTaskHandler(taskHandler *th) {
	mux.HandleFunc(string(model.TaskSendEmail), taskHandler.HandleSendEmail)
}

// ServerConfig configuration options for worker server
type ServerConfig struct {
	AsynqConfig   asynq.Config
	SchedulerOpts *asynq.SchedulerOpts
	MailUtil      mail.Utility
	Limiter       *rate.Limiter
	MailRepo      model.EmailRepository
	SharedCryptor common.SharedCryptor
}

// NewServer return worker server
func NewServer(redisHost string, cfg ServerConfig) (workerPkg.Server, error) {
	srv, err := workerPkg.NewServer(
		redisHost,
		cfg.AsynqConfig,
		cfg.SchedulerOpts,
	)

	th := newTaskHandler(cfg.MailUtil, cfg.Limiter, cfg.MailRepo, cfg.SharedCryptor)

	registerTaskHandler(th)

	return srv, err
}

// Mux return worker mux
func Mux() *asynq.ServeMux {
	return mux
}
