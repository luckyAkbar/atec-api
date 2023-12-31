package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/luckyAkbar/atec-api/internal/common"
	"github.com/luckyAkbar/atec-api/internal/model"
	"github.com/luckyAkbar/atec-api/internal/model/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserRepository_Create(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewUserRepository(kit.DB, mockCacher)
	ctx := context.Background()
	mock := kit.DBmock
	now := time.Now().UTC()
	user := &model.User{
		ID:        uuid.New(),
		Email:     "email",
		Password:  "password",
		Username:  "username",
		IsActive:  false,
		Role:      model.RoleUser,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "users"`).WithArgs(
					user.ID,
					user.Email,
					user.Password,
					user.Username,
					user.IsActive,
					user.Role,
					user.CreatedAt,
					user.UpdatedAt,
					user.DeletedAt,
				).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Create(ctx, user, nil)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err db",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^INSERT INTO "pins"`).WithArgs(
					user.ID,
					user.Email,
					user.Password,
					user.Username,
					user.IsActive,
					user.Role,
					user.CreatedAt,
					user.UpdatedAt,
					user.DeletedAt,
				).WillReturnError(errors.New("db err"))
				mock.ExpectRollback()
			},
			Run: func() {
				err := repo.Create(ctx, user, nil)
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewUserRepository(kit.DB, mockCacher)
	ctx := context.Background()
	mock := kit.DBmock
	id := uuid.New()
	email := "email@gmail.com"

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "users" WHERE`).WithArgs(email).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
			},
			Run: func() {
				user, err := repo.FindByEmail(ctx, email)
				assert.NoError(t, err)
				assert.Equal(t, user.ID, id)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "users" WHERE`).WithArgs(email).WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindByEmail(ctx, email)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "db error",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "users" WHERE`).WithArgs(email).WillReturnError(errors.New("db err"))
			},
			Run: func() {
				_, err := repo.FindByEmail(ctx, email)
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "db err")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}

func TestUserRepository_UpdateActiveStatus(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewUserRepository(kit.DB, mockCacher)
	ctx := context.Background()
	mock := kit.DBmock
	id := uuid.New()
	status := true

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`^UPDATE "users" SET`).
					WithArgs(status, sqlmock.AnyArg(), id).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
				mock.ExpectCommit()
			},
			Run: func() {
				_, err := repo.UpdateActiveStatus(ctx, id, status)
				assert.NoError(t, err)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`^UPDATE "users" SET`).
					WithArgs(status, sqlmock.AnyArg(), id).WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			Run: func() {
				_, err := repo.UpdateActiveStatus(ctx, id, status)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "db err",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`^UPDATE "users" SET`).
					WithArgs(status, sqlmock.AnyArg(), id).WillReturnError(errors.New("err db"))
				mock.ExpectRollback()
			},
			Run: func() {
				_, err := repo.UpdateActiveStatus(ctx, id, status)
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}

func TestUserRepository_CreateChangePasswordSession(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	ctx := context.Background()
	key := "key"
	session := &model.ChangePasswordSession{
		UserID:    uuid.New(),
		ExpiredAt: time.Now().Add(time.Hour * 1).UTC(),
		CreatedAt: time.Now().UTC(),
		CreatedBy: uuid.New(),
	}
	expiry := time.Hour * 1

	r := NewUserRepository(kit.DB, mockCacher)

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mockCacher.EXPECT().Set(ctx, key, session.ToJSONString(), expiry).Times(1).Return(nil)
			},
			Run: func() {
				err := r.CreateChangePasswordSession(ctx, key, expiry, session)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err",
			MockFn: func() {
				mockCacher.EXPECT().Set(ctx, key, session.ToJSONString(), expiry).Times(1).Return(errors.New("redis err"))
			},
			Run: func() {
				err := r.CreateChangePasswordSession(ctx, key, expiry, session)
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewUserRepository(kit.DB, mockCacher)
	ctx := context.Background()
	mock := kit.DBmock
	id := uuid.New()

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "users" WHERE`).WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
			},
			Run: func() {
				user, err := repo.FindByID(ctx, id)
				assert.NoError(t, err)
				assert.Equal(t, user.ID, id)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "users" WHERE`).WithArgs(id).WillReturnError(gorm.ErrRecordNotFound)
			},
			Run: func() {
				_, err := repo.FindByID(ctx, id)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "db error",
			MockFn: func() {
				mock.ExpectQuery(`^SELECT .+ FROM "users" WHERE`).WithArgs(id).WillReturnError(errors.New("db err"))
			},
			Run: func() {
				_, err := repo.FindByID(ctx, id)
				assert.Error(t, err)
				assert.Equal(t, err.Error(), "db err")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}

func TestUserRepository_FindChangePasswordSession(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewUserRepository(kit.DB, mockCacher)
	ctx := context.Background()
	key := "key"
	session := &model.ChangePasswordSession{
		UserID:    uuid.New(),
		ExpiredAt: time.Now().Add(time.Minute * 15).UTC(),
		CreatedAt: time.Now().UTC(),
		CreatedBy: uuid.New(),
	}

	tests := []common.TestStructure{
		{
			Name: "ok found",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, key).Times(1).Return(session.ToJSONString(), nil)
			},
			Run: func() {
				res, err := repo.FindChangePasswordSession(ctx, key)
				assert.NoError(t, err)
				assert.Equal(t, res, session)
			},
		},
		{
			Name: "not found",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, key).Times(1).Return("", redis.Nil)
			},
			Run: func() {
				_, err := repo.FindChangePasswordSession(ctx, key)
				assert.Error(t, err)
				assert.Equal(t, err, ErrNotFound)
			},
		},
		{
			Name: "db err",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, key).Times(1).Return("", errors.New("err db"))
			},
			Run: func() {
				_, err := repo.FindChangePasswordSession(ctx, key)
				assert.Error(t, err)
			},
		},
		{
			Name: "somehow returning invalid json",
			MockFn: func() {
				mockCacher.EXPECT().Get(ctx, key).Times(1).Return("invalid{json}", nil)
			},
			Run: func() {
				_, err := repo.FindChangePasswordSession(ctx, key)
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	mockCacher := mock.NewMockCacher(kit.Ctrl)
	repo := NewUserRepository(kit.DB, mockCacher)
	ctx := context.Background()
	mock := kit.DBmock
	u := &model.User{
		ID:        uuid.New(),
		Email:     "email",
		Password:  "XXXXXXXX",
		Username:  "username",
		IsActive:  true,
		Role:      model.RoleAdmin,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		DeletedAt: gorm.DeletedAt{},
	}

	tests := []common.TestStructure{
		{
			Name: "ok",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET`).
					WithArgs(u.Email, u.Password, u.Username, u.IsActive, u.Role, u.CreatedAt, sqlmock.AnyArg(), u.DeletedAt, u.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Update(ctx, u, nil)
				assert.NoError(t, err)
			},
		},
		{
			Name: "err",
			MockFn: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`^UPDATE "users" SET`).
					WithArgs(u.Email, u.Password, u.Username, u.IsActive, u.Role, u.CreatedAt, sqlmock.AnyArg(), u.DeletedAt, u.ID).
					WillReturnError(errors.New("db err"))
				mock.ExpectCommit()
			},
			Run: func() {
				err := repo.Update(ctx, u, nil)
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}

func TestSDUserRepository_Search(t *testing.T) {
	kit, closer := common.InitializeRepoTestKit(t)
	defer closer()

	repo := NewUserRepository(kit.DB, nil)
	ctx := context.Background()
	mock := kit.DBmock

	tests := []common.TestStructure{
		{
			Name: "ok - include deleted",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "users"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.SearchUserInput{
					IncludeDeleted: true,
				})

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "ok - exclude deleted",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "users" .+ "users"."deleted_at" IS NULL .+`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				res, err := repo.Search(ctx, &model.SearchUserInput{
					IncludeDeleted: false,
				})

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "ok1",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "users" .+ "users"."deleted_at" IS NULL .+`).
					WithArgs(model.RoleAdmin).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				input := &model.SearchUserInput{
					Role:           model.RoleAdmin,
					IncludeDeleted: false,
				}
				res, err := repo.Search(ctx, input)

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "ok2",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "users"`).
					WithArgs("%username%").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				input := &model.SearchUserInput{
					IncludeDeleted: true,
					Username:       "username",
				}
				res, err := repo.Search(ctx, input)

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "ok3",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "users"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			},
			Run: func() {
				input := &model.SearchUserInput{
					IncludeDeleted: true,
					Limit:          10,
				}
				res, err := repo.Search(ctx, input)

				assert.NoError(t, err)
				assert.Equal(t, len(res), 1)
			},
		},
		{
			Name: "db err",
			MockFn: func() {
				mock.ExpectQuery(`SELECT .+ FROM "users"`).
					WillReturnError(errors.New("err db"))
			},
			Run: func() {
				input := &model.SearchUserInput{
					Limit: 10,
				}
				_, err := repo.Search(ctx, input)

				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockFn()
			tt.Run()
		})
	}
}
