package auth

import (
	"log/slog"
	"net/http"

	"github.com/Alhanaqtah/auth/internal/config"
	"github.com/Alhanaqtah/auth/pkg/logger/sl"
	"github.com/go-playground/validator/v10"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type AuthCase interface {
	SignUp(email, password string) error
}

type Controller struct {
	ac     AuthCase
	tCfg   *config.Tokens
	log    *slog.Logger
	valdtr *validator.Validate
}

type Config struct {
	AuthCase     AuthCase
	TokensConfig *config.Tokens
	Log          *slog.Logger
}

type credentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func New(cfg *Config) *Controller {
	return &Controller{
		ac:     cfg.AuthCase,
		tCfg:   cfg.TokensConfig,
		log:    cfg.Log,
		valdtr: validator.New(),
	}
}

func (c *Controller) Register() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/auth", func(r chi.Router) {
		r.Post("/sign-up", c.signUp)
		r.Post("/sign-in", c.signIn)
	})

	return r
}

func (c *Controller) signUp(w http.ResponseWriter, r *http.Request) {
	const op = "controllers.auth.signUp"

	log := c.log.With(
		slog.String("op", op),
		slog.String("req_id", middleware.GetReqID(r.Context())),
	)

	var creds credentials
	if err := render.DecodeJSON(r.Body, &creds); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err), sl.Err(err))
		render.Status(r, http.StatusUnprocessableEntity)
		return
	}

	defer r.Body.Close()

	if err := c.valdtr.Struct(creds); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		render.Status(r, http.StatusBadRequest)
		return
	}

	if err := c.ac.SignUp(creds.Email, creds.Password); err != nil {
		render.Status(r, http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusCreated)
}

func (c *Controller) signIn(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
