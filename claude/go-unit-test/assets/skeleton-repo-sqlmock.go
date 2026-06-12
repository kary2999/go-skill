// 骨架：repo / DAO 层测试（用 sqlmock）
// 用法：复制到 internal/data/<name>_test.go，替换 TODO。
//
// 选型：
//   - sqlmock：快，验证 SQL 字符串 + 参数。PG/MySQL 都支持
//   - 想验证真实 SQL 行为 → 走集成测试（integration.md）
//
// 依赖：github.com/DATA-DOG/go-sqlmock

package data_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestREPO_METHOD(t *testing.T) { // TODO: TestOrderRepo_Create
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		// TODO: 入参字段
		wantErr error
	}{
		{
			name: "happy_path_inserts_and_returns_id",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(
					`INSERT INTO orders (amount, status) VALUES ($1, $2) RETURNING id`,
				)).WithArgs(100, "pending").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name: "db_conn_lost_returns_error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO orders`).
					WillReturnError(errors.New("conn closed"))
				mock.ExpectRollback()
			},
			wantErr: ErrDBUnavailable, // TODO
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer db.Close()

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			// TODO: repo := NewOrderRepo(db)
			var repo interface {
				Create(ctx context.Context, amount int, status string) (int64, error)
			}
			_ = db
			_ = repo

			// TODO: _, err = repo.Create(context.Background(), 100, "pending")
			err = nil

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				require.NoError(t, err)
			}

			// 所有预期必须全被触发
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
