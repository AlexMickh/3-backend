package user_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestSaveUser(t *testing.T) {
	type fields struct {
		db Postgres
	}
	type args struct {
		ctx      context.Context
		email    string
		phone    string
		password string
	}

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

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
				ctx:      context.Background(),
				email:    "example@mail.com",
				phone:    "89543476748",
				password: "password123",
			},
			wantErr:     nil,
			wantMockErr: nil,
		},
		{
			name: "already exists case",
			fields: fields{
				db: mock,
			},
			args: args{
				ctx:      context.Background(),
				email:    "example@mail.com",
				phone:    "89543476748",
				password: "password123",
			},
			wantErr: errs.ErrUserAlreadyExists,
			wantMockErr: &pgconn.PgError{
				Code: "23505",
			},
		},
		{
			name: "unexpected error case",
			fields: fields{
				db: mock,
			},
			args: args{
				ctx:      context.Background(),
				email:    "example@mail.com",
				phone:    "89543476748",
				password: "password123",
			},
			wantErr:     unexpectedErr,
			wantMockErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.wantMockErr == nil {
				id, _ := uuid.NewV7()
				row := mock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(tt.args.email, tt.args.phone, tt.args.password).
					WillReturnRows(row)
			} else {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs(tt.args.email, tt.args.phone, tt.args.password).
					WillReturnError(tt.wantMockErr)
			}

			u := &UserRepository{
				db: tt.fields.db,
			}
			id, err := u.SaveUser(tt.args.ctx, tt.args.email, tt.args.phone, tt.args.password)
			require.ErrorIs(t, err, tt.wantErr)

			_, err = uuid.Parse(id.String())
			require.NoError(t, err)
		})
	}
}
