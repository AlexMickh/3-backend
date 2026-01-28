package user_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_SaveUser(t *testing.T) {
	type args struct {
		ctx      context.Context
		email    string
		password string
	}

	someError := errors.New("some error")

	tests := []struct {
		name      string
		args      args
		want      int64
		wantDbErr error
		wantErr   error
	}{
		{
			name: "good case",
			args: args{
				ctx:      context.Background(),
				email:    "example@mail.com",
				password: "some password",
			},
			want:      1,
			wantDbErr: nil,
			wantErr:   nil,
		},
		{
			name: "user already exists case",
			args: args{
				ctx:      context.Background(),
				email:    "example@mail.com",
				password: "some password",
			},
			want:      1,
			wantDbErr: sqlite3.Error{ExtendedCode: sqlite3.ErrConstraintUnique},
			wantErr:   errs.ErrUserAlreadyExists,
		},
		{
			name: "unexpected error case",
			args: args{
				ctx:      context.Background(),
				email:    "example@mail.com",
				password: "some password",
			},
			want:      1,
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
				row := sqlmock.NewRows([]string{"id"}).AddRow(tt.want)
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(tt.args.email, tt.args.password).
					WillReturnRows(row)
			} else {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(tt.args.email, tt.args.password).
					WillReturnError(tt.wantDbErr)
			}

			u := &UserRepository{
				db: db,
			}
			got, err := u.SaveUser(tt.args.ctx, tt.args.email, tt.args.password)
			require.ErrorIs(t, err, tt.wantErr)

			if err == nil {
				require.Equal(t, tt.want, got)
			}
		})
	}
}
