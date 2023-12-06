package console

import (
	"context"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/db"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/repository"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/sweet-go/stdlib/encryption"
	"golang.org/x/crypto/bcrypt"
)

var adminCMD = &cobra.Command{
	Use:  "admin",
	Long: "create user with role admin",
	Run:  adminFn,
}

func init() {
	adminCMD.PersistentFlags().String("email", "", "email")
	_ = adminCMD.MarkPersistentFlagRequired("email")

	adminCMD.PersistentFlags().String("password", "", "password")
	_ = adminCMD.MarkPersistentFlagRequired("password")

	adminCMD.PersistentFlags().String("username", "", "username")
	_ = adminCMD.MarkPersistentFlagRequired("username")
	RootCMD.AddCommand(adminCMD)
}

func adminFn(cmd *cobra.Command, _ []string) {
	validation := validator.New()

	username := cmd.Flag("username").Value.String()
	email := cmd.Flag("email").Value.String()
	password := cmd.Flag("password").Value.String()

	if err := validation.Var(email, "required,email"); err != nil {
		logrus.Error("flag email must be a valid email address")
		os.Exit(1)
	}

	if len(username) < 3 {
		logrus.Error("flag username must be at least 3 characters")
		os.Exit(1)
	}

	if len(password) < 8 {
		logrus.Error("flag password must be at least 8 characters")
		os.Exit(1)
	}

	key, err := encryption.ReadKeyFromFile("./private.pem")
	if err != nil {
		panic(err)
	}

	sharedCryptor := common.NewSharedCryptor(&common.CreateCryptorOpts{
		HashCost:      bcrypt.DefaultCost,
		EncryptionKey: key.Bytes,
		IV:            "4e6064d3814c2cd22c550155655fefc6", //4e6064d3814c2cd22c550155655fefc6
		BlockSize:     common.DefaultBlockSize,
	})

	db.InitializePostgresConn()

	userRepo := repository.NewUserRepository(db.PostgresDB, nil)

	emailEnc, err := sharedCryptor.Encrypt(email)
	if err != nil {
		logrus.WithError(err).Error("failed to encrypt email")
		os.Exit(1)
	}

	pwEnc, err := sharedCryptor.Hash([]byte(password))
	if err != nil {
		logrus.WithError(err).Error("failed to encrypt email")
		os.Exit(1)
	}

	now := time.Now().UTC()

	admin := &model.User{
		ID:        uuid.New(),
		Email:     emailEnc,
		Password:  pwEnc,
		Username:  username,
		IsActive:  true,
		Role:      model.RoleAdmin,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := userRepo.Create(context.Background(), admin, nil); err != nil {
		logrus.WithError(err).Error("failed to save admin data to db")
		os.Exit(1)
	}

	os.Exit(1)
}
