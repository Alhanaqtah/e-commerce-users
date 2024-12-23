package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"e-commerce-users/internal/config"
	http_lib "e-commerce-users/internal/lib/http"
	"e-commerce-users/internal/services"
	"e-commerce-users/pkg/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type AuthService interface {
	SignUp(ctx context.Context, name, surname, birthdate, email, password string) error
	SignIn(ctx context.Context, name, password string) (string, string, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	Confirm(ctx context.Context, email, code string) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	ResendCode(ctx context.Context, email string) error
}

type Controller struct {
	as     AuthService
	tCfg   *config.Tokens
	valdtr *validator.Validate
}

type Config struct {
	AuthService AuthService
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

type tokensRequest struct {
	AccessToken  string `json:"access_token" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type tokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type confirmRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required"`
}

type resendRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func New(cfg *Config) *Controller {
	return &Controller{
		as:     cfg.AuthService,
		tCfg:   cfg.TknsCfg,
		valdtr: validator.New(),
	}
}

func (c *Controller) Register() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/sign-up", c.signUp)
	r.Post("/sign-in", c.signIn)
	r.Post("/logout", c.logout)
	r.Post("/confirm", c.confirm)
	r.Post("/resend", c.resend)
	r.Post("/refresh", c.refresh)

	return r
}

func (c *Controller) signUp(w http.ResponseWriter, r *http.Request) {
	const op = "http.auth.signUp"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var creds signUpCredentials
	if err := render.DecodeJSON(r.Body, &creds); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http_lib.ErrUnprocessableEntity(w, r)
		return
	}

	defer r.Body.Close() //nolint:errcheck

	if err := c.valdtr.Struct(creds); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http_lib.ErrInvalid(w, r, err)
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
			http_lib.ErrConflict(w, r, "User already exists")
			return
		}

		http_lib.ErrInternal(w, r)
		return
	}

	log.Info("user signed up succesfully", slog.String("email", creds.Email))

	render.Status(r, http.StatusCreated)
	render.Render(w, r, http_lib.RespOk("Registration successful. A confirmation code has been sent to your email")) //nolint:errcheck
}

func (c *Controller) signIn(w http.ResponseWriter, r *http.Request) {
	const op = "http.auth.signUp"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var creds signInCredentials
	if err := render.DecodeJSON(r.Body, &creds); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http_lib.ErrUnprocessableEntity(w, r)
		return
	}

	defer r.Body.Close() //nolint:errcheck

	if err := c.valdtr.Struct(creds); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http_lib.ErrInvalid(w, r, err)
		return
	}

	assessToken, refreshToken, err := c.as.SignIn(r.Context(), creds.Email, creds.Password)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http_lib.ErrUnauthorized(w, r, "User not found")
			return
		}
		if errors.Is(err, services.ErrInvalidCredentials) {
			http_lib.ErrUnauthorized(w, r, "Invalid credentials")
			return
		}

		http_lib.ErrInternal(w, r)
		return
	}

	log.Info("user logged in successfully", slog.String("email", creds.Email))

	render.Status(r, http.StatusOK)
	render.Render(w, r, tokensResponse{ //nolint:errcheck
		AccessToken:  assessToken,
		RefreshToken: refreshToken,
	})
}

func (c *Controller) logout(w http.ResponseWriter, r *http.Request) {
	const op = "http.auth.logout"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var tkns tokensRequest
	if err := render.DecodeJSON(r.Body, &tkns); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http_lib.ErrUnprocessableEntity(w, r)
		return
	}

	if err := c.valdtr.Struct(tkns); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http_lib.ErrInvalid(w, r, err)
		return
	}

	if err := c.as.Logout(r.Context(), tkns.AccessToken, tkns.RefreshToken); err != nil {
		if errors.Is(err, services.ErrTokenInvalid) {
			http_lib.ErrUnauthorized(w, r, "Invalid token")
			return
		}

		if errors.Is(err, services.ErrTokenBlacklisted) {
			http_lib.ErrUnauthorized(w, r, "Token revoked")
			return
		}

		http_lib.ErrInternal(w, r)
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, http_lib.RespOk("User logged out succesfully")) //nolint:errcheck
}

func (c *Controller) confirm(w http.ResponseWriter, r *http.Request) {
	const op = "http.auth.confirm"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var confirmReq confirmRequest
	if err := render.DecodeJSON(r.Body, &confirmReq); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http_lib.ErrUnprocessableEntity(w, r)
		return
	}

	defer r.Body.Close() //nolint:errcheck

	if err := c.valdtr.Struct(confirmReq); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http_lib.ErrInvalid(w, r, err)
		return
	}

	accTkn, rfrshTkn, err := c.as.Confirm(r.Context(), confirmReq.Email, confirmReq.Code)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http_lib.ErrUnauthorized(w, r, "User not found")
			return
		}
		if errors.Is(err, services.ErrCode) {
			http_lib.ErrBadRequest(w, r)
			return
		}

		http_lib.ErrInternal(w, r)
		return
	}

	render.JSON(w, r, tokensResponse{
		AccessToken:  accTkn,
		RefreshToken: rfrshTkn,
	})
}

func (c *Controller) resend(w http.ResponseWriter, r *http.Request) {
	const op = "http.auth.resend"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var rsntCode resendRequest
	if err := render.DecodeJSON(r.Body, &rsntCode); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http_lib.ErrUnprocessableEntity(w, r)
		return
	}

	defer r.Body.Close() //nolint:errcheck

	if err := c.valdtr.Struct(rsntCode); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http_lib.ErrInvalid(w, r, err)
		return
	}

	if err := c.as.ResendCode(r.Context(), rsntCode.Email); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http_lib.ErrUnauthorized(w, r, "User with given email does't exists")
			return
		}
		if errors.Is(err, services.ErrNoActionRequired) {
			http_lib.ErrConflict(w, r, "User already active")
			return
		}

		http_lib.ErrInternal(w, r)
	}

	render.JSON(w, r, http_lib.RespOk("Confirmation code sent to your email address"))
}

func (c *Controller) refresh(w http.ResponseWriter, r *http.Request) {
	const op = "http.auth.refresh"

	log := http_lib.GetCtxLogger(r.Context())
	log = log.With(slog.String("op", op))

	var rfrshReq refreshRequest
	if err := render.DecodeJSON(r.Body, &rfrshReq); err != nil {
		log.Debug("failed to parse JSON", sl.Err(err))
		http_lib.ErrUnprocessableEntity(w, r)
		return
	}

	defer r.Body.Close() //nolint:errcheck

	if err := c.valdtr.Struct(rfrshReq); err != nil {
		log.Error("some fields are invalid", sl.Err(err))
		http_lib.ErrInvalid(w, r, err)
		return
	}

	accTkn, rfrshTkn, err := c.as.Refresh(r.Context(), rfrshReq.RefreshToken)
	if err != nil {
		if errors.Is(err, services.ErrTokenBlacklisted) {
			http_lib.ErrUnauthorized(w, r, "Token revoked")
			return
		}
		if errors.Is(err, services.ErrTokenExpired) {
			http_lib.ErrUnauthorized(w, r, "Token expired")
			return
		}
		if errors.Is(err, services.ErrUnexpectedTokenType) {
			http_lib.ErrUnauthorized(w, r, "Unexpected token type: expected refresh token")
			return
		}

		http_lib.ErrInternal(w, r)
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, tokensResponse{ //nolint:errcheck
		AccessToken:  accTkn,
		RefreshToken: rfrshTkn,
	})
}

func (t tokensResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
