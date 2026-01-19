package jwt

import (
	"fmt"
	"testing"
	"time"

	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJwtManager_NewJwt(t *testing.T) {
	type fields struct {
		secret string
		jwtTtl time.Duration
	}
	type args struct {
		userID uuid.UUID
		role   models.UserRole
	}

	id, err := uuid.NewV7()
	require.NoError(t, err)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "good case",
			fields: fields{
				secret: "some secret",
				jwtTtl: 5 * time.Minute,
			},
			args: args{
				userID: id,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JwtManager{
				secret: tt.fields.secret,
				jwtTtl: tt.fields.jwtTtl,
			}
			got, err := j.NewJwt(tt.args.userID, tt.args.role)
			require.ErrorIs(t, err, tt.wantErr)
			fmt.Println(got)
		})
	}
}
