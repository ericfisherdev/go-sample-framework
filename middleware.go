package finishline

import (
	"github.com/justinas/nosurf"
	"net/http"
	"strconv"
)

func (f *FinishLine) SessionLoad(next http.Handler) http.Handler {
	return f.Session.LoadAndSave(next)
}

func (f *FinishLine) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(f.config.cookie.secure)

	csrfHandler.ExemptGlob("/api/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Domain:   f.config.cookie.domain,
	})

	return csrfHandler
}
