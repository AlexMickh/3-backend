package token_repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSaveToken(t *testing.T) {
	type args struct {
		ctx   context.Context
		token models.Token
	}

	someError := errors.New("some error")

	tests := []struct {
		name      string
		args      args
		wantDbErr error
		wantErr   error
	}{
		{
			name: "good case",
			args: args{
				ctx: context.Background(),
				token: models.Token{
					Token:     "some token",
					UserID:    1,
					Type:      models.TokenTypeValidateEmail,
					ExpiresAt: time.Now().Add(5 * time.Minute),
				},
			},
			wantDbErr: nil,
			wantErr:   nil,
		},
		{
			name: "unexpected error case",
			args: args{
				ctx: context.Background(),
				token: models.Token{
					Token:     "some token",
					UserID:    1,
					Type:      models.TokenTypeValidateEmail,
					ExpiresAt: time.Now().Add(5 * time.Minute),
				},
			},
			wantDbErr: someError,
			wantErr:   someError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			if tt.wantDbErr == nil {
				mock.ExpectExec("INSERT INTO tokens").
					WithArgs(tt.args.token.Token, tt.args.token.UserID, string(tt.args.token.Type), tt.args.token.ExpiresAt).
					WillReturnResult(sqlmock.NewResult(1, 1))
			} else {
				mock.ExpectExec("INSERT INTO tokens").
					WithArgs(tt.args.token.Token, tt.args.token.UserID, string(tt.args.token.Type), tt.args.token.ExpiresAt).
					WillReturnError(tt.wantDbErr)
			}

			tr := &TokenRepository{
				db: db,
			}
			err = tr.SaveToken(tt.args.ctx, tt.args.token)
			require.ErrorIs(t, err, tt.wantErr)

			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}
