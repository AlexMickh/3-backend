package auth_service

import (
	"context"
	"errors"
	"testing"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	type args struct {
		ctx context.Context
		req dtos.RegisterDto
	}

	var id int64 = 1
	tokenErr := errors.New("token error")

	tests := []struct {
		name                string
		args                args
		want                int64
		wantErr             error
		wantUserMockErr     error
		wantUserMockReturn  int64
		wantTokenMockErr    error
		wantTokenMockReturn models.Token
	}{
		{
			name: "good case",
			args: args{
				ctx: context.Background(),
				req: dtos.RegisterDto{
					Email:    "example@email.com",
					Password: "12345",
				},
			},
			want:                id,
			wantErr:             nil,
			wantUserMockErr:     nil,
			wantUserMockReturn:  id,
			wantTokenMockErr:    nil,
			wantTokenMockReturn: models.Token{Token: "some token"},
		},
		{
			name: "too long password case",
			args: args{
				ctx: context.Background(),
				req: dtos.RegisterDto{
					Email:    "example@email.com",
					Password: "111111111111111111111111111111111111111111111111111111111111111111111111111111111111111",
				},
			},
			want:                0,
			wantErr:             bcrypt.ErrPasswordTooLong,
			wantUserMockErr:     nil,
			wantUserMockReturn:  id,
			wantTokenMockErr:    nil,
			wantTokenMockReturn: models.Token{},
		},
		{
			name: "user service error case",
			args: args{
				ctx: context.Background(),
				req: dtos.RegisterDto{
					Email:    "example@email.com",
					Password: "1",
				},
			},
			want:                0,
			wantErr:             errs.ErrUserAlreadyExists,
			wantUserMockErr:     errs.ErrUserAlreadyExists,
			wantUserMockReturn:  0,
			wantTokenMockErr:    nil,
			wantTokenMockReturn: models.Token{},
		},
		{
			name: "token error case",
			args: args{
				ctx: context.Background(),
				req: dtos.RegisterDto{
					Email:    "example@email.com",
					Password: "12345",
				},
			},
			want:                0,
			wantErr:             tokenErr,
			wantUserMockErr:     nil,
			wantUserMockReturn:  id,
			wantTokenMockErr:    tokenErr,
			wantTokenMockReturn: models.Token{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userService := NewMockUserService(t)
			userService.EXPECT().CreateUser(
				mock.AnythingOfType("context.backgroundCtx"),
				mock.AnythingOfType("string"),
				mock.AnythingOfType("string"),
			).Return(tt.wantUserMockReturn, tt.wantUserMockErr).Maybe()

			tokenService := NewMockTokenService(t)
			tokenService.EXPECT().CreateToken(
				mock.AnythingOfType("context.backgroundCtx"),
				mock.AnythingOfType("int64"),
				mock.AnythingOfType("models.TokenType"),
			).Return(tt.wantTokenMockReturn, tt.wantTokenMockErr).Maybe()

			a := &AuthService{
				userService:  userService,
				emailQueue:   make(chan [2]string, 5),
				tokenService: tokenService,
			}

			got, err := a.Register(tt.args.ctx, tt.args.req)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.want, got)

			if err == nil {
				message := <-a.emailQueue
				require.Equal(t, tt.args.req.Email, message[0])
				require.Equal(t, tt.wantTokenMockReturn.Token, message[1])
			}
		})
	}
}
