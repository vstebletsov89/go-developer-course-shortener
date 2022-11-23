package postgres

import (
	"context"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/repository/postgres/testhelpers"
	"go-developer-course-shortener/internal/app/types"
	"go-developer-course-shortener/internal/configs"
	"log"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestStorage interface {
	repository.Repository
}

type StorageTestSuite struct {
	suite.Suite
	TestStorage
	container *testhelpers.TestDatabase
}

func (sts *StorageTestSuite) SetupTest() {

	storageContainer := testhelpers.NewTestDatabase(sts.T())

	dsn := storageContainer.ConnectionString(sts.T())
	log.Printf("DATABASE_DSN: %v", dsn)
	sts.T().Setenv("DATABASE_DSN", dsn)

	config, err := configs.ReadConfig()

	conn, err := pgx.Connect(context.Background(), config.DatabaseDsn)
	require.NoError(sts.T(), err)

	storage, err := NewDBRepository(conn)
	require.NoError(sts.T(), err)

	sts.TestStorage = storage
	sts.container = storageContainer
}

func (sts *StorageTestSuite) TearDownTest() {
	sts.container.Close(sts.T())
}

func TestStorageTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	suite.Run(t, new(StorageTestSuite))
}

func (sts *StorageTestSuite) Test_Ping() {
	tests := []struct {
		name    string
		wantRes bool
	}{
		{
			name:    "positive test",
			wantRes: true,
		},
	}
	for _, tt := range tests {
		sts.Run(tt.name, func() {
			s := sts.TestStorage
			if res := s.Ping(); res != tt.wantRes {
				sts.T().Errorf("Ping() error = %v, wantErr %v", res, tt.wantRes)
			}
		})
	}
}

func (sts *StorageTestSuite) TestDBRepository_SaveURL() {
	tests := []struct {
		name        string
		userID      string
		shortURL    string
		originalURL string
		wantErr     bool
	}{
		{
			name:        "positive test",
			userID:      "s_user",
			shortURL:    "s_short",
			originalURL: "s_orig",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		sts.Run(tt.name, func() {
			s := sts.TestStorage
			if err := s.SaveURL(tt.userID, tt.shortURL, tt.originalURL); (err != nil) != tt.wantErr {
				sts.T().Errorf("SaveURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func (sts *StorageTestSuite) TestDBRepository_GetShortURLByOriginalURL() {
	tests := []struct {
		name        string
		userID      string
		shortURL    string
		originalURL string
		wantErr     bool
	}{
		{
			name:        "positive test",
			userID:      "sbo_user",
			shortURL:    "sbo_short1",
			originalURL: "sbo_orig1",
			wantErr:     false,
		},
		{
			name:        "negative test",
			userID:      "",
			shortURL:    "",
			originalURL: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		sts.Run(tt.name, func() {
			s := sts.TestStorage
			if !tt.wantErr {
				if err := s.SaveURL(tt.userID, tt.shortURL, tt.originalURL); (err != nil) != tt.wantErr {
					sts.T().Errorf("SaveURL() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			got, err := s.GetShortURLByOriginalURL(tt.originalURL)
			if (err != nil) != tt.wantErr {
				sts.T().Errorf("GetShortURLByOriginalURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.shortURL {
				sts.T().Errorf("GetShortURLByOriginalURL() got = %v, want %v", got, tt.shortURL)
			}
		})
	}
}

func (sts *StorageTestSuite) TestDBRepository_GetUserStorage() {
	tests := []struct {
		name    string
		userID  string
		links   types.BatchLinks
		want    []types.Link
		wantErr bool
	}{
		{
			name:   "positive test",
			userID: "us_user",
			links: types.BatchLinks{types.BatchLink{
				CorrelationID: "us_id1",
				ShortURL:      "us_short1",
				OriginalURL:   "us_orig1",
			},
				types.BatchLink{
					CorrelationID: "us_id2",
					ShortURL:      "us_short2",
					OriginalURL:   "us_orig2",
				},
			},
			want: []types.Link{
				{
					ShortURL:    "us_short1",
					OriginalURL: "us_orig1",
				},
				{
					ShortURL:    "us_short2",
					OriginalURL: "us_orig2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		sts.Run(tt.name, func() {
			s := sts.TestStorage
			_, err := s.SaveBatchURLS(tt.userID, tt.links)
			if (err != nil) != tt.wantErr {
				sts.T().Errorf("SaveBatchURLS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := s.GetUserStorage(tt.userID)
			if (err != nil) != tt.wantErr {
				sts.T().Errorf("GetUserStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				sts.T().Errorf("GetUserStorage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func (sts *StorageTestSuite) TestDBRepository_DeleteURLS() {
	tests := []struct {
		name      string
		userID    string
		ctx       context.Context
		links     types.BatchLinks
		shortURLS []string
		wantErr   bool
	}{
		{
			name:   "positive test",
			userID: "d_user",
			ctx:    context.Background(),
			links: types.BatchLinks{types.BatchLink{
				CorrelationID: "d_id1",
				ShortURL:      "d_short1",
				OriginalURL:   "d_orig1",
			},
				types.BatchLink{
					CorrelationID: "d_id2",
					ShortURL:      "d_short2",
					OriginalURL:   "d_orig2",
				},
			},
			shortURLS: []string{"d_short1", "d_short2"},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		sts.Run(tt.name, func() {
			s := sts.TestStorage
			if !tt.wantErr {
				_, err := s.SaveBatchURLS(tt.userID, tt.links)
				if (err != nil) != tt.wantErr {
					sts.T().Errorf("SaveBatchURLS() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			if err := s.DeleteURLS(tt.ctx, tt.userID, tt.shortURLS); (err != nil) != tt.wantErr {
				sts.T().Errorf("DeleteURLS() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func (sts *StorageTestSuite) TestDBRepository_SaveBatchURLS() {
	tests := []struct {
		name    string
		userID  string
		links   types.BatchLinks
		want    types.ResponseBatch
		wantErr bool
	}{
		{
			name:   "positive test",
			userID: "sb_user",
			links: types.BatchLinks{types.BatchLink{
				CorrelationID: "sb_id1",
				ShortURL:      "sb_short1",
				OriginalURL:   "sb_orig1",
			},
				types.BatchLink{
					CorrelationID: "sb_id2",
					ShortURL:      "sb_short2",
					OriginalURL:   "sb_orig2",
				},
			},
			want: types.ResponseBatch{
				types.ResponseBatchJSON{
					CorrelationID: "sb_id1",
					ShortURL:      "sb_short1",
				},
				types.ResponseBatchJSON{
					CorrelationID: "sb_id2",
					ShortURL:      "sb_short2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		sts.Run(tt.name, func() {
			s := sts.TestStorage
			got, err := s.SaveBatchURLS(tt.userID, tt.links)
			if (err != nil) != tt.wantErr {
				sts.T().Errorf("SaveBatchURLS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				sts.T().Errorf("SaveBatchURLS() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func (sts *StorageTestSuite) TestDBRepository_GetURL() {
	tests := []struct {
		name        string
		userID      string
		shortURL    string
		originalURL string
		wantURL     types.OriginalLink
		wantErr     bool
	}{
		{
			name:        "positive test",
			userID:      "g_user",
			shortURL:    "g_short",
			originalURL: "g_orig",
			wantURL:     types.OriginalLink{OriginalURL: "g_orig"},
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		sts.Run(tt.name, func() {
			s := sts.TestStorage

			if err := s.SaveURL(tt.userID, tt.shortURL, tt.originalURL); (err != nil) != tt.wantErr {
				sts.T().Errorf("SaveURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := s.GetURL(tt.shortURL)
			if (err != nil) != tt.wantErr {
				sts.T().Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantURL) {
				sts.T().Errorf("GetURL() got = %v, want %v", got, tt.wantURL)
			}
		})
	}
}

func (sts *StorageTestSuite) Test_Negative() {
	tests := []struct {
		name      string
		links     types.BatchLinks
		shortURLS []string
		wantRes   bool
		wantErr   bool
	}{
		{
			name: "negative tests for all methods",
			links: types.BatchLinks{types.BatchLink{
				CorrelationID: "neg_id1",
				ShortURL:      "neg_short1",
				OriginalURL:   "neg_orig1",
			},
				types.BatchLink{
					CorrelationID: "neg_id2",
					ShortURL:      "neg_short2",
					OriginalURL:   "neg_orig2",
				},
			},
			shortURLS: []string{"neg_short1", "neg_short2"},
			wantRes:   false,
			wantErr:   true,
		},
	}

	sts.TestStorage.ReleaseStorage() // close db connection

	for _, tt := range tests {
		sts.Run(tt.name, func() {
			s := sts.TestStorage
			if res := s.Ping(); res != tt.wantRes {
				sts.T().Errorf("Ping() error = %v, wantRes %v", res, tt.wantRes)
			}

			if err := s.SaveURL("", "", ""); (err != nil) != tt.wantErr {
				sts.T().Errorf("SaveURL() error = %v, wantErr %v", err, tt.wantErr)
			}

			_, err := s.GetURL("")
			if (err != nil) != tt.wantErr {
				sts.T().Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			_, err = s.GetUserStorage("")
			if (err != nil) != tt.wantErr {
				sts.T().Errorf("GetUserStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			_, err = s.GetShortURLByOriginalURL("")
			if (err != nil) != tt.wantErr {
				sts.T().Errorf("GetShortURLByOriginalURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			_, err = s.SaveBatchURLS("", tt.links)
			if (err != nil) != tt.wantErr {
				sts.T().Errorf("SaveBatchURLS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := s.DeleteURLS(context.Background(), "", tt.shortURLS); (err != nil) != tt.wantErr {
				sts.T().Errorf("DeleteURLS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
