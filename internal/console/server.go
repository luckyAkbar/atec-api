package console

import (
	"context"
	"crypto"
	"crypto/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/config"
	"github.com/luckyAkbar/atec-api/internal/db"
	"github.com/luckyAkbar/atec-api/internal/delivery/rest"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/luckyAkbar/atec-api/internal/usecase"
	"github.com/luckyAkbar/atec-api/internal/worker"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/sweet-go/stdlib/encryption"
	stdhttp "github.com/sweet-go/stdlib/http"
	workerPkg "github.com/sweet-go/stdlib/worker"
	"golang.org/x/crypto/bcrypt"
)

var serverCMD = &cobra.Command{
	Use:  "server",
	Long: "run the API server",
	Run:  serverFn,
}

func init() {
	RootCMD.AddCommand(serverCMD)
}

func serverFn(_ *cobra.Command, _ []string) {
	key, err := encryption.ReadKeyFromFile("./private.pem")
	if err != nil {
		panic(err)
	}

	fontBytes, err := os.ReadFile("./assets/font.ttf")
	if err != nil {
		panic(err)
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:         config.RedisAddr(),
		Password:     config.RedisPassword(),
		DB:           config.RedisCacheDB(),
		MinIdleConns: config.RedisMinIdleConn(),
		MaxIdleConns: config.RedisMaxIdleConn(),
	})

	sharedCryptor := common.NewSharedCryptor(&common.CreateCryptorOpts{
		HashCost:      bcrypt.DefaultCost,
		EncryptionKey: key.Bytes,
		IV:            config.IVKey(),
		BlockSize:     common.DefaultBlockSize,
	})

	apirespGen := stdhttp.NewStandardAPIResponseGenerator(&encryption.SignOpts{
		Random:  rand.Reader,
		PrivKey: key.PrivateKey,
		Alg:     crypto.SHA256,
		PSSOpts: nil,
	})

	db.InitializePostgresConn()
	cacher := db.NewCacher(redisClient)

	userRepo := repository.NewUserRepository(db.PostgresDB, cacher)
	pinRepo := repository.NewPinRepository(db.PostgresDB)
	emailRepo := repository.NewEmailRepository(db.PostgresDB)
	accessTokenRepo := repository.NewAccessTokenRepository(db.PostgresDB)
	sdtemplateRepo := repository.NewSDTemplateRepository(db.PostgresDB)
	sdpackageRepo := repository.NewSDPackageRepository(db.PostgresDB)
	sdtRepo := repository.NewSDTestResultRepository(db.PostgresDB)

	workerPkgClient, err := workerPkg.NewClient(config.WorkerBrokerHost())
	if err != nil {
		panic(err)
	}

	workerClient := worker.NewClient(workerPkgClient)

	emailUsecase := usecase.NewEmailUsecase(emailRepo, workerClient, sharedCryptor)
	userUsecase := usecase.NewUserUsecase(userRepo, pinRepo, sharedCryptor, emailUsecase, accessTokenRepo, db.PostgresDB)
	authUsecase := usecase.NewAuthUsecase(accessTokenRepo, userRepo, sharedCryptor)
	sdtemplateUsecase := usecase.NewSDTemplateUsecase(sdtemplateRepo)
	sdpackageUsecase := usecase.NewSDPackageUsecase(sdpackageRepo, sdtemplateRepo)
	sdtUsecase := usecase.NewSDTestResultUsecase(sdtRepo, sdpackageRepo, sharedCryptor, db.PostgresDB, f)

	httpServer := echo.New()

	httpServer.Pre(middleware.AddTrailingSlash())
	httpServer.Use(middleware.Logger())
	httpServer.Use(middleware.Recover())
	httpServer.Use(middleware.CORS())

	rootGroup := httpServer.Group("")

	rest.NewService(rootGroup, apirespGen, userUsecase, authUsecase, sdtemplateUsecase, sdpackageUsecase, sdtUsecase)

	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error, 1)
	quitCh := make(chan bool, 1)
	signal.Notify(sigCh, os.Interrupt)

	go func() {
		for {
			select {
			case <-sigCh:
				logrus.Info("shutting down the server")
				gracefulShutdown(httpServer)
				quitCh <- true
			case e := <-errCh:
				logrus.Error(e)
				gracefulShutdown(httpServer)
				quitCh <- true
			}
		}
	}()

	// Start server
	go func() {
		if err := httpServer.Start(config.ServerPort()); err != nil && err != http.ErrServerClosed {
			httpServer.Logger.Fatal("shutting down the server: ", err.Error())
		}
	}()

	<-quitCh
	logrus.Info("exiting")
}

func gracefulShutdown(srv *echo.Echo) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		srv.Logger.Fatal(err)
	}
}
