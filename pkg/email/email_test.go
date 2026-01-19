package email

import (
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/k1LoW/smtptest"
	"github.com/stretchr/testify/require"
)

func TestSendWelcomeMessages(t *testing.T) {
	type fields struct {
		cfg               EmailConfig
		auth              smtp.Auth
		welcomeEmailQueue chan [2]string
	}
	type args struct {
		ctx  context.Context
		tmpl *template.Template
	}

	ts, auth, err := smtptest.NewServerWithAuth()
	require.NoError(t, err)
	t.Cleanup(func() {
		ts.Close()
	})

	tmpl, err := template.ParseFiles("./templates/verify-email.html")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	code := "12345"

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "good case",
			fields: fields{
				cfg: EmailConfig{
					Host:     ts.Host,
					Port:     ts.Port,
					FromAddr: "test@mail.com",
					Password: "",
				},
				auth:              auth,
				welcomeEmailQueue: make(chan [2]string, 5),
			},
			args: args{
				ctx:  ctx,
				tmpl: tmpl,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &EmailService{
				cfg:               tt.fields.cfg,
				auth:              tt.fields.auth,
				welcomeEmailQueue: tt.fields.welcomeEmailQueue,
			}
			e.welcomeEmailQueue <- [2]string{"test2@mail.com", code}
			t.Log("test starts")
			go func() {
				time.Sleep(200 * time.Millisecond)
				cancel()
			}()
			e.sendWelcomeMessages(tt.args.ctx, tt.args.tmpl)
			require.Equal(t, 1, len(ts.Messages()))

			buf := make([]byte, 512)
			ts.Messages()[0].Body.Read(buf)
			require.True(t, strings.Contains(string(buf), fmt.Sprintf("<a href=\"http://localhost:8080/auth/verify/%s\">link</a>", code)))
		})
	}
}
