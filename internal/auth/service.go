package auth

import (
	"crypto/subtle"
	"net/http"
	"time"

	authsvc "github.com/go-pkgz/auth/v2"
	"github.com/go-pkgz/auth/v2/middleware"
	"github.com/go-pkgz/auth/v2/provider"
	authjwt "github.com/go-pkgz/auth/v2/token"

	"gobkd/internal/config"
)

type Service struct {
	service    *authsvc.Service
	authRoutes http.Handler
	authMW     middleware.Authenticator
}

func New(cfg config.Config) *Service {
	svc := authsvc.NewService(authsvc.Opts{
		SecretReader: authjwt.SecretFunc(func(_ string) (string, error) {
			return cfg.AuthSecret, nil
		}),
		TokenDuration:  time.Hour,
		CookieDuration: 24 * time.Hour,
		Issuer:         "gobkd",
		URL:            cfg.AppBaseURL,
		SecureCookies:  cfg.AppEnv == "prod",
		DisableXSRF:    cfg.AppEnv != "prod",
		SameSiteCookie: http.SameSiteLaxMode,
	})

	svc.AddDirectProvider("local", provider.CredCheckerFunc(func(user, password string) (bool, error) {
		userOK := subtle.ConstantTimeCompare([]byte(user), []byte(cfg.AuthLocalUser)) == 1
		passOK := subtle.ConstantTimeCompare([]byte(password), []byte(cfg.AuthLocalPass)) == 1
		return userOK && passOK, nil
	}))

	authRoutes, _ := svc.Handlers()

	return &Service{
		service:    svc,
		authRoutes: authRoutes,
		authMW:     svc.Middleware(),
	}
}

func (s *Service) Routes() http.Handler {
	return s.authRoutes
}

func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return s.authMW.Auth(next)
}

func (s *Service) Trace(next http.Handler) http.Handler {
	return s.authMW.Trace(next)
}

func (s *Service) CurrentUser(r *http.Request) (authjwt.User, error) {
	return authjwt.GetUserInfo(r)
}
