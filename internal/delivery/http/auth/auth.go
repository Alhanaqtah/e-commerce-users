package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"e-commerce-users/internal/config"
	"e-commerce-users/internal/services"
	"e-commerce-users/pkg/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type AuthService interface {
	SignUp(ctx context.Context, name, surname, birthdate, email, password string) error
	SignIn(ctx context.Context, name, password string) (string, string, error)
}

type Controller struct {
	as     AuthService
	tCfg   *config.Tokens
	log    *slog.Logger
	valdtr *validator.Validate
}

type Config struct {
	AuthService AuthService
	Log         *slog.Logger
	TknsCfg     *config.Tokens
}

type signUpCredentials struct {
	Name      string    `json:"name" validate:"required"`
	Surname   string    `json:"surname" validate:"required"`
	Birthdate time.Time `json:"birthdate" validate:"required"`
	Email     string    `json:"email" validate:"required,email"`
	Password  string    `json:"password" validate:"required"`
}

type signInCredentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func New(cfg *Config) *Controller {
	return &Controller{
		as:     cfg.AuthService,
		log:    cfg.Log,
		tCfg:   cfg.TknsCfg,
		valdtr: validator.New(),
	}
}

func (c *Controller) Register() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/sign-up", c.signUp)
	r.Post("/sign-in", c.signIn)

	return r
}

func (c *Controller) signUp(w http.ResponseWriter, r *http.Request) {
	const op = "controllers.auth.signUp"

	log := c.log.With(
		slog.String("op", op),
	)

	var creds signUpCredentials
	defer r.Body.Close()
	if err := render.DecodeJSON(r.Body, &creds); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := c.valdtr.Struct(creds); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := c.as.SignUp(r.Context(),
		creds.Name,
		creds.Surname,
		creds.Birthdate.Format("2006-01-02"),
		creds.Email,
		creds.Password,
	); err != nil {
		if errors.Is(err, services.ErrExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusCreated)
}

func (c *Controller) signIn(w http.ResponseWriter, r *http.Request) {
	const op = "controllers.auth.signUp"

	log := c.log.With(
		slog.String("op", op),
	)

	var creds signInCredentials
	defer r.Body.Close()
	if err := render.DecodeJSON(r.Body, &creds); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := c.valdtr.Struct(creds); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	assessToken, refreshToken, err := c.as.SignIn(r.Context(), creds.Email, creds.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, tokens{
		AccessToken:  assessToken,
		RefreshToken: refreshToken,
	})
}

func (t tokens) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
