package cookies

import (
	"net/http"
	"time"
)

func Set(w http.ResponseWriter, name string, value string, ttl time.Duration) {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(ttl),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}

	http.SetCookie(w, &cookie)
}
