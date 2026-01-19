package token_repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestSaveToken(t *testing.T) {
	type fields struct {
		db Postgres
	}
	type args struct {
		ctx   context.Context
		token models.Token
	}

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID, err := uuid.NewV7()
	require.NoError(t, err)

	unexpectedErr := errors.New("some error")

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     error
		wantMockErr error
	}{
		{
			name: "good case",
			fields: fields{
				db: mock,
			},
			args: args{
				ctx: context.Background(),
				token: models.Token{
					Token:     "some token",
					UserID:    userID,
					Type:      models.TokenTypeValidateEmail,
					ExpiresAt: time.Now().Add(5 * time.Minute),
				},
			},
			wantErr:     nil,
			wantMockErr: nil,
		},
		{
			name: "unexpected error case",
			fields: fields{
				db: mock,
			},
			args: args{
				ctx: context.Background(),
				token: models.Token{
					Token:     "some token",
					UserID:    userID,
					Type:      models.TokenTypeValidateEmail,
					ExpiresAt: time.Now().Add(5 * time.Minute),
				},
			},
			wantErr:     unexpectedErr,
			wantMockErr: unexpectedErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantMockErr == nil {
				mock.ExpectExec("INSERT INTO tokens").
					WithArgs(tt.args.token.Token, tt.args.token.UserID, string(tt.args.token.Type), tt.args.token.ExpiresAt).
					WillReturnResult(pgconn.NewCommandTag("good"))
			} else {
				mock.ExpectExec("INSERT INTO tokens").
					WithArgs(tt.args.token.Token, tt.args.token.UserID, string(tt.args.token.Type), tt.args.token.ExpiresAt).
					WillReturnError(tt.wantMockErr)
			}

			tr := &TokenRepository{
				db: tt.fields.db,
			}
			err := tr.SaveToken(tt.args.ctx, tt.args.token)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
