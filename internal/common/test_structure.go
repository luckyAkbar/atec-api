// Package common will hold usefull data, struct etc that commonly used within the code
package common

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestStructure represent common unit test structure
type TestStructure struct {
	Name   string
	MockFn func()
	Run    func()
}

// RepoTestKit represent repo test kit, it will hold mock and db instance for repo test case
// It will also hold gomock controller for mocking
// It will also hold close function for closing db connection
// It will be used by repository layer test case
type RepoTestKit struct {
	DBmock sqlmock.Sqlmock
	DB     *gorm.DB
	Ctrl   *gomock.Controller
}

// InitializeRepoTestKit will initialize repo test kit, it will return RepoTestKit and close function
func InitializeRepoTestKit(t *testing.T) (kit *RepoTestKit, close func()) {
	dbconn, dbmock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: dbconn}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	ctrl := gomock.NewController(t)

	tk := &RepoTestKit{
		Ctrl:   ctrl,
		DBmock: dbmock,
		DB:     gormDB,
	}

	return tk, func() {
		if conn, _ := tk.DB.DB(); conn != nil {
			_ = conn.Close()
		}
	}
}
