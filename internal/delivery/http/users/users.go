package users

import (
	"context"
	"log/slog"
	"net/http"

	"e-commerce-users/internal/config"
	http_lib "e-commerce-users/internal/lib/http"
	"e-commerce-users/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
)

type UsersService interface {
	GetProfile(ctx context.Context, id string) (*models.User, error)
}

type Controller struct {
	us      UsersService
	tknsCfg config.Tokens
}

type Config struct {
	UsrSrvc UsersService
	TknsCfg config.Tokens
}

func New(cfg *Config) *Controller {
	return &Controller{
		us:      cfg.UsrSrvc,
		tknsCfg: cfg.TknsCfg,
	}
}

func (c *Controller) Register() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/me", func(r chi.Router) {
		r.Use(jwtauth.Verifier(
			jwtauth.New("HS256", []byte(c.tknsCfg.Secret), nil),
		))
		r.Use(http_lib.Authenticator)

		r.Get("/", c.getProfile)
	})

	return r
}

func (c *Controller) getProfile(w http.ResponseWriter, r *http.Request) {
	const op = "controllers.users.getProfile"

	log := http_lib.GetCtxLogger(r.Context())
	log.With(slog.String("op", op))

	ctx := r.Context()

	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http_lib.ErrInternal(w, r)
		return
	}

	id := claims["sub"].(string)

	user, err := c.us.GetProfile(ctx, id)
	if err != nil {
		http_lib.ErrInternal(w, r)
		return
	}

	render.JSON(w, r, user)
}
