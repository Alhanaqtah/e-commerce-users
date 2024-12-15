package auth_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"e-commerce-users/internal/config"
	auth_ctrl "e-commerce-users/internal/delivery/http/auth"
	auth_mock "e-commerce-users/internal/delivery/http/auth/mock"
	http_lib "e-commerce-users/internal/lib/http"
	"e-commerce-users/pkg/logger/handlers/slogdiscard"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestController_signUp(t *testing.T) {
	// Init deps
	authSrvc := new(auth_mock.AuthService)

	r := chi.NewRouter()
	ctrl := auth_ctrl.New(
		&auth_ctrl.Config{
			AuthService: authSrvc,
			TknsCfg: &config.Tokens{
				Secret:     "secret",
				AccessTTL:  5 * time.Minute,
				RefreshTTL: 15 * time.Minute,
			},
		},
	)

	logger := slogdiscard.NewDiscardLogger()

	r.Use(http_lib.Logging(logger))

	r.Mount("/auth", ctrl.Register())

	tests := []struct {
		name                 string
		inputBody            string
		expectedStatus       int
		expectedResponseBody string
		mockBehavior         func()
	}{
		{
			name:                 "Correct input",
			inputBody:            `{"name": "Jhon", "surname": "Doe", "birthdate": "2000-01-01T00:00:00Z", "email": "jhon@mail.com", "password": "qwerty"}`,
			expectedStatus:       http.StatusCreated,
			expectedResponseBody: `{"status": "Ok", "message": "Registration successful. A confirmation code has been sent to your email"}`,
			mockBehavior: func() {
				authSrvc.On("SignUp",
					mock.Anything,
					"Jhon",
					"Doe",
					"2000-01-01",
					"jhon@mail.com",
					"qwerty",
				).Return(nil)
			},
		},
		{
			name:                 "Empty body",
			inputBody:            ``,
			expectedStatus:       http.StatusUnprocessableEntity,
			expectedResponseBody: `{"status": "Error","message": "Unprocessable entity"}`,
			mockBehavior:         func() {},
		},
		{
			name:           "Invalid body",
			inputBody:      `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedResponseBody: `
			{
				"status": "Error",
				"message": "Some fields are invalid",
				"errors": {
					"birthdate": "field must satisfy 'required' constraint",
					"email": "field must satisfy 'required' constraint",
					"name": "field must satisfy 'required' constraint",
					"password": "field must satisfy 'required' constraint",
					"surname": "field must satisfy 'required' constraint"
				}
			}`,
			mockBehavior: func() {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBufferString(tc.inputBody))
			w := httptest.NewRecorder()
			defer w.Result().Body.Close()

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code) //nolint:bodyclose
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}

func TestController_signIn(t *testing.T) {
	authSrvc := new(auth_mock.AuthService)

	r := chi.NewRouter()
	ctrl := auth_ctrl.New(
		&auth_ctrl.Config{
			AuthService: authSrvc,
			TknsCfg: &config.Tokens{
				Secret:     "secret",
				AccessTTL:  5 * time.Minute,
				RefreshTTL: 15 * time.Minute,
			},
		},
	)

	logger := slogdiscard.NewDiscardLogger()

	r.Use(http_lib.Logging(logger))

	r.Mount("/auth", ctrl.Register())

	tests := []struct {
		name                 string
		inputBody            string
		expectedStatus       int
		expectedResponseBody string
		mockBehavior         func()
	}{
		{
			name:           "Correct input",
			inputBody:      `{"email": "jhon@mail.com", "password": "qwerty"}`,
			expectedStatus: http.StatusOK,
			expectedResponseBody: `
			{
 			   "access_token": "new-access-token",
    			"refresh_token": "new-refresh-token"
			}
			`,
			mockBehavior: func() {
				authSrvc.On(
					"SignIn",
					mock.Anything,
					"jhon@mail.com",
					"qwerty",
				).Return(
					"new-access-token",
					"new-refresh-token",
					nil,
				)
			},
		},
		{
			name:                 "Empty body",
			inputBody:            ``,
			expectedStatus:       http.StatusUnprocessableEntity,
			expectedResponseBody: `{"status": "Error","message": "Unprocessable entity"}`,
			mockBehavior:         func() {},
		},
		{
			name:           "Invalid body",
			inputBody:      `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedResponseBody: `
			{
				"status": "Error",
				"message": "Some fields are invalid",
				"errors": {
					"email": "field must satisfy 'required' constraint",
					"password": "field must satisfy 'required' constraint"
				}
			}`,
			mockBehavior: func() {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBufferString(tc.inputBody))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Result().StatusCode) //nolint:bodyclose
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}

func TestController_logout(t *testing.T) {
	authSrvc := new(auth_mock.AuthService)

	r := chi.NewRouter()
	ctrl := auth_ctrl.New(
		&auth_ctrl.Config{
			AuthService: authSrvc,
			TknsCfg: &config.Tokens{
				Secret:     "secret",
				AccessTTL:  5 * time.Minute,
				RefreshTTL: 15 * time.Minute,
			},
		},
	)

	logger := slogdiscard.NewDiscardLogger()

	r.Use(http_lib.Logging(logger))

	r.Mount("/auth", ctrl.Register())

	tests := []struct {
		name                 string
		inputBody            string
		expectedStatus       int
		expectedResponseBody string
		mockBehavior         func()
	}{
		{
			name:           "Correct input",
			inputBody:      `{"access_token": "access-token", "refresh_token": "refresh-token"}`,
			expectedStatus: http.StatusOK,
			expectedResponseBody: `
			{
				"status": "Ok",
				"message": "User logged out succesfully"
			}`,
			mockBehavior: func() {
				authSrvc.On("Logout",
					mock.Anything,
					"access-token",
					"refresh-token",
				).Return(nil)
			},
		},
		{
			name:                 "Empty body",
			inputBody:            ``,
			expectedStatus:       http.StatusUnprocessableEntity,
			expectedResponseBody: `{"status": "Error","message": "Unprocessable entity"}`,
			mockBehavior:         func() {},
		},
		{
			name:           "Invalid body",
			inputBody:      `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedResponseBody: `
			{
				"status": "Error",
				"message": "Some fields are invalid",
				"errors": {
					"accesstoken": "field must satisfy 'required' constraint",
					"refreshtoken": "field must satisfy 'required' constraint"
				}
			}`,
			mockBehavior: func() {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest("POST", "/auth/logout", bytes.NewBufferString(tc.inputBody))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Result().StatusCode) //nolint:bodyclose
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}

func TestController_confirm(t *testing.T) {
	authSrvc := new(auth_mock.AuthService)

	r := chi.NewRouter()
	ctrl := auth_ctrl.New(
		&auth_ctrl.Config{
			AuthService: authSrvc,
			TknsCfg: &config.Tokens{
				Secret:     "secret",
				AccessTTL:  5 * time.Minute,
				RefreshTTL: 15 * time.Minute,
			},
		},
	)

	logger := slogdiscard.NewDiscardLogger()

	r.Use(http_lib.Logging(logger))

	r.Mount("/auth", ctrl.Register())

	tests := []struct {
		name                 string
		inputBody            string
		expectedStatus       int
		expectedResponseBody string
		mockBehavior         func()
	}{
		{
			name:           "Correct input",
			inputBody:      `{"email": "jhon@mail.com","code": "YwhpHx"}`,
			expectedStatus: http.StatusOK,
			expectedResponseBody: `
			{
				"access_token": "new-access-token",
				"refresh_token": "new-refresh-token"
			}`,
			mockBehavior: func() {
				authSrvc.On("Confirm",
					mock.Anything,
					"jhon@mail.com",
					"YwhpHx",
				).Return(
					"new-access-token",
					"new-refresh-token",
					nil,
				)
			},
		},
		{
			name:                 "Empty body",
			inputBody:            ``,
			expectedStatus:       http.StatusUnprocessableEntity,
			expectedResponseBody: `{"status": "Error","message": "Unprocessable entity"}`,
			mockBehavior:         func() {},
		},
		{
			name:           "Invalid body",
			inputBody:      `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedResponseBody: `
			{
				"status": "Error",
				"message": "Some fields are invalid",
				"errors": {
					"email": "field must satisfy 'required' constraint",
					"code": "field must satisfy 'required' constraint"
				}
			}`,
			mockBehavior: func() {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest("POST", "/auth/confirm", bytes.NewBufferString(tc.inputBody))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Result().StatusCode) //nolint:bodyclose
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}

func TestController_resend(t *testing.T) {
	authSrvc := new(auth_mock.AuthService)

	r := chi.NewRouter()
	ctrl := auth_ctrl.New(
		&auth_ctrl.Config{
			AuthService: authSrvc,
			TknsCfg: &config.Tokens{
				Secret:     "secret",
				AccessTTL:  5 * time.Minute,
				RefreshTTL: 15 * time.Minute,
			},
		},
	)

	logger := slogdiscard.NewDiscardLogger()

	r.Use(http_lib.Logging(logger))

	r.Mount("/auth", ctrl.Register())

	tests := []struct {
		name                 string
		inputBody            string
		expectedStatus       int
		expectedResponseBody string
		mockBehavior         func()
	}{
		{
			name:           "Correct input",
			inputBody:      `{"email": "jhon@mail.com"}`,
			expectedStatus: http.StatusOK,
			expectedResponseBody: `
			{
				"status": "Ok",
				"message": "Confirmation code sent to your email address"
			}`,
			mockBehavior: func() {
				authSrvc.On("ResendCode",
					mock.Anything,
					"jhon@mail.com",
				).Return(nil)
			},
		},
		{
			name:                 "Empty body",
			inputBody:            ``,
			expectedStatus:       http.StatusUnprocessableEntity,
			expectedResponseBody: `{"status": "Error","message": "Unprocessable entity"}`,
			mockBehavior:         func() {},
		},
		{
			name:           "Invalid body",
			inputBody:      `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedResponseBody: `
			{
				"status": "Error",
				"message": "Some fields are invalid",
				"errors": {
					"email": "field must satisfy 'required' constraint"
				}
			}`,
			mockBehavior: func() {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest("POST", "/auth/resend", bytes.NewBufferString(tc.inputBody))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Result().StatusCode) //nolint:bodyclose
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}

func TestController_refresh(t *testing.T) {
	authSrvc := new(auth_mock.AuthService)

	r := chi.NewRouter()
	ctrl := auth_ctrl.New(
		&auth_ctrl.Config{
			AuthService: authSrvc,
			TknsCfg: &config.Tokens{
				Secret:     "secret",
				AccessTTL:  5 * time.Minute,
				RefreshTTL: 15 * time.Minute,
			},
		},
	)

	logger := slogdiscard.NewDiscardLogger()

	r.Use(http_lib.Logging(logger))

	r.Mount("/auth", ctrl.Register())

	tests := []struct {
		name                 string
		inputBody            string
		expectedStatus       int
		expectedResponseBody string
		mockBehavior         func()
	}{
		{
			name:           "Correct input",
			inputBody:      `{"refresh_token": "refresh-token"}`,
			expectedStatus: http.StatusOK,
			expectedResponseBody: `
			{
				"access_token": "new-access-token",
				"refresh_token": "new-refresh-token"
			}`,
			mockBehavior: func() {
				authSrvc.On("Refresh",
					mock.Anything,
					"refresh-token",
				).Return(
					"new-access-token",
					"new-refresh-token",
					nil,
				)
			},
		},
		{
			name:                 "Empty body",
			inputBody:            ``,
			expectedStatus:       http.StatusUnprocessableEntity,
			expectedResponseBody: `{"status": "Error","message": "Unprocessable entity"}`,
			mockBehavior:         func() {},
		},
		{
			name:           "Invalid body",
			inputBody:      `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedResponseBody: `
			{
				"status": "Error",
				"message": "Some fields are invalid",
				"errors": {
					"refreshtoken": "field must satisfy 'required' constraint"
				}
			}`,
			mockBehavior: func() {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBufferString(tc.inputBody))
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Result().StatusCode) //nolint:bodyclose
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}
