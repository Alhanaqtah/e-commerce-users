package users_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"e-commerce-users/internal/config"
	"e-commerce-users/internal/delivery/http/users"
	users_mock "e-commerce-users/internal/delivery/http/users/mock"
	http_lib "e-commerce-users/internal/lib/http"
	"e-commerce-users/internal/models"
	"e-commerce-users/pkg/logger/handlers/slogdiscard"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestController_getProfile(t *testing.T) {
	usrsSrvc := new(users_mock.UsersService)

	r := chi.NewRouter()
	ctrl := users.New(
		&users.Config{
			UsrSrvc: usrsSrvc,
			TknsCfg: config.Tokens{
				Secret:     "secret",
				AccessTTL:  5 * time.Minute,
				RefreshTTL: 15 * time.Minute,
			},
		},
	)

	logger := slogdiscard.NewDiscardLogger()
	r.Use(http_lib.Logging(logger))
	r.Mount("/users", ctrl.Register())

	tests := []struct {
		name                 string
		inputToken           string
		expectedStatus       int
		expectedResponseBody string
		mockBehavior         func()
	}{
		{
			name:           "Valid Token and User Found",
			inputToken:     `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiY3VzdG9tZXIiLCJzdWIiOiIzZjc4YWM3Mi0zN2MxLTQ3ZWUtOTc0Ny1iYjA2MjE0ZjUzMTAiLCJ0eXBlIjoiYWNjZXNzIiwidmVyc2lvbiI6MX0.QFriLPs-aGG-fFpN0LAvSJ4xej5hMbTGq7BIArBO0o0`,
			expectedStatus: http.StatusOK,
			expectedResponseBody: `
			{
				"id": "3f78ac72-37c1-47ee-9747-bb06214f5310",
				"name": "Jhon",
				"surname": "Doe",
				"birthdate": "2006-03-03T00:00:00Z",
				"role": "customer",
				"email": "jhon@mail.com",
				"created_at": "2024-12-11T12:51:55.935946Z"
			}`,
			mockBehavior: func() {
				birthdate, _ := time.Parse(time.RFC3339, "2006-03-03T00:00:00Z")
				createdAt, _ := time.Parse(time.RFC3339, "2024-12-11T12:51:55.935946Z")

				usrsSrvc.On(
					"GetProfile",
					mock.Anything,
					"3f78ac72-37c1-47ee-9747-bb06214f5310",
				).Return(&models.User{
					ID:        "3f78ac72-37c1-47ee-9747-bb06214f5310",
					Name:      "Jhon",
					Surname:   "Doe",
					Birthdate: birthdate,
					Role:      "customer",
					Email:     "jhon@mail.com",
					CreatedAt: createdAt,
				}, nil)
			},
		},
		{
			name:           "Invalid Token",
			inputToken:     `invalid.token.here`,
			expectedStatus: http.StatusUnauthorized,
			expectedResponseBody: `
			{
				"status": "Error",
				"message": "Token is unauthorized"
			}`,
			mockBehavior: func() {},
		},
		{
			name:           "User Not Found",
			inputToken:     `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiY3VzdG9tZXIiLCJzdWIiOiJpbmNvcnJlY3QtaWQiLCJ0eXBlIjoiYWNjZXNzIiwidmVyc2lvbiI6MX0.jsdNIRN5Y135xE8YA3D4McsWLlwex2F4eNCnPviDbfo`,
			expectedStatus: http.StatusInternalServerError,
			expectedResponseBody: `
			{
				"status": "Error",
				"message": "Internal error"
			}`,
			mockBehavior: func() {
				usrsSrvc.On(
					"GetProfile",
					mock.Anything,
					"incorrect-id",
				).Return(nil, fmt.Errorf("user not found"))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			req := httptest.NewRequest("GET", "/users/me", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.inputToken))

			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedResponseBody, w.Body.String())
		})
	}
}
