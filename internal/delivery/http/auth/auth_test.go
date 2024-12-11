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
			expectedResponseBody: `{
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

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Result().StatusCode)
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}

}
