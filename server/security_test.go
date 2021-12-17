package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/frain-dev/convoy/mocks"
	"github.com/go-chi/chi/v5"

	"github.com/frain-dev/convoy/config"
	"github.com/golang/mock/gomock"
)

func TestApplicationHandler_CreateAPIKey(t *testing.T) {
	var app *applicationHandler

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	groupRepo := mocks.NewMockGroupRepository(ctrl)
	appRepo := mocks.NewMockApplicationRepository(ctrl)
	eventRepo := mocks.NewMockEventRepository(ctrl)
	eventDeliveryRepo := mocks.NewMockEventDeliveryRepository(ctrl)
	eventQueue := mocks.NewMockQueuer(ctrl)
	apiKeyRepo := mocks.NewMockAPIKeyRepo(ctrl)

	app = newApplicationHandler(eventRepo, eventDeliveryRepo, appRepo, groupRepo, apiKeyRepo, eventQueue)

	tt := []struct {
		name           string
		cfgPath        string
		statusCode     int
		stripTimestamp bool
		body           *strings.Reader
		dbFn           func(app *applicationHandler)
	}{
		{
			name:           "create api key",
			stripTimestamp: true,
			cfgPath:        "./testdata/Auth_Config/no-auth-convoy.json",
			statusCode:     http.StatusCreated,
			body: strings.NewReader(`{
					"key": "12344",
					"expires_at": "2022-01-02T15:04:05+01:00",
                    "role": {
                        "type": "admin",
                        "groups": [
                            "sendcash-pay"
                        ]
                    }
                }`),
			dbFn: func(app *applicationHandler) {
				a, _ := app.apiKeyRepo.(*mocks.MockAPIKeyRepo)
				a.EXPECT().CreateAPIKey(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
		},
		{
			name:           "create api key without key field",
			stripTimestamp: true,
			cfgPath:        "./testdata/Auth_Config/no-auth-convoy.json",
			statusCode:     http.StatusCreated,
			body: strings.NewReader(`{
					"expires_at": "2022-01-02T15:04:05+01:00", 
					"role": {
                        "type": "ui_admin",
                        "groups": [
                            "sendcash-pay"
                        ]
                    }
                }`),
			dbFn: func(app *applicationHandler) {
				a, _ := app.apiKeyRepo.(*mocks.MockAPIKeyRepo)
				a.EXPECT().CreateAPIKey(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
		},
		{
			name:           "invalid expiry date",
			stripTimestamp: false,
			cfgPath:        "./testdata/Auth_Config/no-auth-convoy.json",
			statusCode:     http.StatusBadRequest,
			body: strings.NewReader(`{
					"expires_at": "2020-01-02T15:04:05+01:00", 
					"role": {
                        "type": "ui_admin",
                        "groups": [
                            "sendcash-pay"
                        ]
                    }
                }`),
			dbFn: nil,
		},
		{
			name:           "create api key without expires_at field",
			stripTimestamp: true,
			cfgPath:        "./testdata/Auth_Config/no-auth-convoy.json",
			statusCode:     http.StatusCreated,
			body: strings.NewReader(`{
					"role": {
                        "type": "ui_admin",
                        "groups": [
                            "sendcash-pay"
                        ]
                    }
                }`),
			dbFn: func(app *applicationHandler) {
				a, _ := app.apiKeyRepo.(*mocks.MockAPIKeyRepo)
				a.EXPECT().CreateAPIKey(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
		},
		{
			name:           "invalid role",
			stripTimestamp: false,
			cfgPath:        "./testdata/Auth_Config/no-auth-convoy.json",
			statusCode:     http.StatusBadRequest,
			body: strings.NewReader(`{
					"key": "12344",
					"expires_at": "2022-01-02T15:04:05+01:00",
                    "role": {
                        "type": "admin",
                        "groups": []
                    }
                }`),
			dbFn: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			url := "/api/v1/security/keys"
			req := httptest.NewRequest(http.MethodPost, url, tc.body)
			req.SetBasicAuth("test", "test")
			w := httptest.NewRecorder()
			rctx := chi.NewRouteContext()
			req.Header.Add("Content-Type", "application/json")

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Arrange Expectations
			if tc.dbFn != nil {
				tc.dbFn(app)
			}

			err := config.LoadConfig(tc.cfgPath)
			if err != nil {
				t.Errorf("Failed to load config file: %v", err)
			}
			initRealmChain(t)

			router := buildRoutes(app)

			// Act
			router.ServeHTTP(w, req)

			// Assert
			if w.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, w.Code)
			}

			if tc.stripTimestamp {
				d := stripTimestamp(t, "apiKey", w.Body)
				w.Body = d
			}

			verifyMatch(t, *w)
		})
	}

}
