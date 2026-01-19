package cash

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

type testObject struct {
	expiresAt time.Time
}

func (t testObject) ExpiresAt() time.Time {
	return t.expiresAt
}

func TestNew(t *testing.T) {
	type args struct {
		gcCallPeriod time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "good case",
			args: args{
				gcCallPeriod: 2 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				goleak.VerifyNone(t)
			})

			ctx, cancel := context.WithTimeout(t.Context(), tt.args.gcCallPeriod+200*time.Millisecond)
			defer cancel()

			cash := New[string, testObject](ctx, tt.args.gcCallPeriod)

			cash.data["test"] = testObject{
				expiresAt: time.Now().Add(tt.args.gcCallPeriod - time.Second),
			}

			time.Sleep(tt.args.gcCallPeriod + 500*time.Millisecond)

			require.Equal(t, 0, len(cash.data))
		})
	}
}
