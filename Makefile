SHELL:=/bin/bash

ifdef test_run
	TEST_ARGS := -run $(test_run)
endif

run_command=go run main.go server

migrate_up=go run main.go migrate --direction=up --step=0
migrate_down=go run main.go migrate --direction=down --step=0

dependency:
	@echo ">> Downloading Dependencies"
	@go mod download

swag-init:
	@echo ">> Running swagger init"
	@swag init --parseDependency --parseInternal

check-modd-exists:
	@modd --version > /dev/null	

run: check-modd-exists
	@modd -f ./.modd/server.modd.conf

worker: check-modd-exists
	@modd -f ./.modd/worker.modd.conf

lint: check-cognitive-complexity
	golangci-lint run --print-issued-lines=false --exclude-use-default=false --enable=revive --enable=goimports  --enable=unconvert --enable=unparam --concurrency=2

check-gotest:
ifeq (, $(shell which richgo))
	$(warning "richgo is not installed, falling back to plain go test")
	$(eval TEST_BIN=go test)
else
	$(eval TEST_BIN=richgo test)
endif

ifdef test_run
	$(eval TEST_ARGS := -run $(test_run))
endif
	$(eval test_command=$(TEST_BIN) ./... $(TEST_ARGS) --cover)

test-only: check-gotest mockgen
	SVC_DISABLE_CACHING=true $(test_command)

test: lint test-only

check-cognitive-complexity:
	find . -type f -name '*.go' -not -name "*.pb.go" -not -name "mock*.go" -not -name "generated.go" -not -name "federation.go" \
      -exec gocognit -over 15 {} +

migrate:
	@if [ "$(DIRECTION)" = "" ] || [ "$(STEP)" = "" ]; then\
    	$(migrate_up);\
	else\
		go run main.go migrate --direction=$(DIRECTION) --step=$(STEP);\
    fi

internal/model/mock/mock_email_usecase.go:
	mockgen -destination=internal/model/mock/mock_email_usecase.go -package=mock github.com/luckyAkbar/atec-api/internal/model EmailUsecase

internal/model/mock/mock_email_repository.go:
	mockgen -destination=internal/model/mock/mock_email_repository.go -package=mock github.com/luckyAkbar/atec-api/internal/model EmailRepository

internal/model/mock/mock_worker_client.go:
	mockgen -destination=internal/model/mock/mock_worker_client.go -package=mock github.com/luckyAkbar/atec-api/internal/model WorkerClient

internal/model/mock/mock_pin_repository.go:
	mockgen -destination=internal/model/mock/mock_pin_repository.go -package=mock github.com/luckyAkbar/atec-api/internal/model PinRepository

internal/model/mock/mock_user_usecase.go:
	mockgen -destination=internal/model/mock/mock_user_usecase.go -package=mock github.com/luckyAkbar/atec-api/internal/model UserUsecase

internal/model/mock/mock_user_repository.go:
	mockgen -destination=internal/model/mock/mock_user_repository.go -package=mock github.com/luckyAkbar/atec-api/internal/model UserRepository

internal/common/mock/mock_shared_cryptor.go:
	mockgen -destination=internal/common/mock/mock_shared_cryptor.go -package=mock github.com/luckyAkbar/atec-api/internal/common SharedCryptor

internal/common/mock/mock_access_token_repository.go:
	mockgen -destination=internal/model/mock/mock_access_token_repository.go -package=mock github.com/luckyAkbar/atec-api/internal/model AccessTokenRepository

internal/common/mock/mock_auth_usecase.go:
	mockgen -destination=internal/model/mock/mock_auth_usecase.go -package=mock github.com/luckyAkbar/atec-api/internal/model AuthUsecase

internal/common/mock/mock_cacher.go:
	mockgen -destination=internal/model/mock/mock_cacher.go -package=mock github.com/luckyAkbar/atec-api/internal/model Cacher

mockgen: clean \
	internal/model/mock/mock_email_usecase.go \
	internal/model/mock/mock_email_repository.go \
	internal/model/mock/mock_worker_client.go \
	internal/model/mock/mock_pin_repository.go \
	internal/model/mock/mock_user_usecase.go \
	internal/model/mock/mock_user_repository.go \
	internal/common/mock/mock_shared_cryptor.go \
	internal/common/mock/mock_access_token_repository.go \
	internal/common/mock/mock_auth_usecase.go \
	internal/common/mock/mock_cacher.go

clean:
	find -type f -name 'mock_*.go' -delete