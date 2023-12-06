// Package config holds all the function to read a specific configuration value
package config

import (
	// this blank import is used to run the init function of this stdlib/config package
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/sendinblue/APIv3-go-library/lib"
	"github.com/spf13/viper"

	// used to indirectly call the init function of this stdlib/config package
	_ "github.com/sweet-go/stdlib/config"
)

// PostgresDSN returns postgres DSN
func PostgresDSN() string {
	host := viper.GetString("postgres.host")
	db := viper.GetString("postgres.db")
	user := viper.GetString("postgres.user")
	pw := viper.GetString("postgres.pw")
	port := viper.GetString("postgres.port")
	sslMode := viper.GetString("postgres.ssl_mode")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, pw, db, port, sslMode)
}

// WorkerBrokerHost returns worker broker host
func WorkerBrokerHost() string {
	return viper.GetString("worker.broker_host")
}

// WorkerConcurency returns worker concurrency. Default to 10
func WorkerConcurency() int {
	num := viper.GetInt("worker.concurrency")
	if num <= 0 {
		return 10
	}

	return num
}

// WorkerLogLevel returns worker log level. Default to info
func WorkerLogLevel() asynq.LogLevel {
	level := viper.GetString("worker.log_level")
	switch strings.ToUpper(level) {
	default:
		return asynq.InfoLevel
	case "DEBUG":
		return asynq.DebugLevel
	case "INFO":
		return asynq.InfoLevel
	case "WARN":
		return asynq.WarnLevel
	case "ERROR":
		return asynq.ErrorLevel
	case "FATAL":
		return asynq.FatalLevel
	}
}

// WorkerLimiterRetryInterval returns worker retry interval in seconds
func WorkerLimiterRetryInterval() time.Duration {
	return viper.GetDuration("worker.limiter.retry_interval_seconds")
}

// WorkerLimiterLimit returns worker limiter limit
func WorkerLimiterLimit() int {
	return viper.GetInt("worker.limiter.limit")
}

// WorkerLimiterBurst returns worker limiter burst
func WorkerLimiterBurst() int {
	return viper.GetInt("worker.limiter.burst")
}

// Env returns application environment
func Env() string {
	return viper.GetString("env")
}

// LogLevel returns application log level
func LogLevel() string {
	return viper.GetString("server.log.level")
}

// ServerPort returns application server port
func ServerPort() string {
	return fmt.Sprintf(":%s", viper.GetString("server.port"))
}

// SendinblueAPIKey get API key for send in blue
func SendinblueAPIKey() string {
	return viper.GetString("sendinblue.api_key")
}

// SendInBlueSender generate sendinblue sender using configured sender name and sender email
func SendInBlueSender() *lib.SendSmtpEmailSender {
	return &lib.SendSmtpEmailSender{
		Name:  viper.GetString("sendinblue.sender_name"),
		Email: viper.GetString("sendinblue.sender_email"),
	}
}

// SendInBlueIsActivated is activated sendinblue
func SendInBlueIsActivated() bool {
	return viper.GetBool("sendinblue.is_activated")
}

// MailgunIsActivated is activated mailgun
func MailgunIsActivated() bool {
	return viper.GetBool("mailgun.is_activated")
}

// MailgunDomain mailgun domain
func MailgunDomain() string {
	return viper.GetString("mailgun.domain")
}

// MailgunPrivateAPIKey mailgun private api key
func MailgunPrivateAPIKey() string {
	return viper.GetString("mailgun.private_api_key")
}

// MailgunPublicAPIKey mailgun public api key
func MailgunPublicAPIKey() string {
	return viper.GetString("mailgun.public_api_key")
}

// MailgunSenderEmail mailgun sender email address, shown in the receipient as the sender email
func MailgunSenderEmail() string {
	return viper.GetString("mailgun.sender_email")
}

// PinExpiryMinutes return pin expiry in minutes. Default to 5 minutes
func PinExpiryMinutes() int {
	minutes := viper.GetInt("server.pin.expiry_minutes")
	if minutes <= 0 {
		return 5
	}

	return minutes
}

// PinMaxRetry return max retry for pin validation. Default to 3
func PinMaxRetry() int {
	tries := viper.GetInt("server.pin.max_tries")
	if tries <= 0 {
		return 3
	}

	return tries
}

// AccessTokenActiveDuration returns access token active duration. Default to 1 hour
func AccessTokenActiveDuration() time.Duration {
	minutes := viper.GetInt("server.auth.access_token_duration_minutes")
	if minutes <= 0 {
		return time.Minute * 60
	}

	return time.Minute * time.Duration(minutes)
}

// ChangePasswordBaseURL return change password base url. Should point to FE page and immediately check the session validity
func ChangePasswordBaseURL() string {
	return viper.GetString("server.user.change_password_base_url")
}

// RedisAddr redis address
func RedisAddr() string {
	return viper.GetString("redis.addr")
}

// RedisPassword redis password
func RedisPassword() string {
	return viper.GetString("redis.password")
}

// RedisCacheDB redis db
func RedisCacheDB() int {
	return viper.GetInt("redis.db")
}

// RedisMinIdleConn min idle
func RedisMinIdleConn() int {
	return viper.GetInt("redis.min")
}

// RedisMaxIdleConn max idle
func RedisMaxIdleConn() int {
	return viper.GetInt("redis.max")
}
