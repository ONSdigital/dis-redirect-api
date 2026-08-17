package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-redirect-api/api"
	"github.com/ONSdigital/dis-redirect-api/config"
	"github.com/ONSdigital/dis-redirect-api/store"
	authorisation "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testRedirectAPIURL = "http://localhost:29900"
)

func TestSetupPrivateEndpoints(t *testing.T) {
	Convey("Given an API instance with private endpoints enabled", t, func() {
		dataStore := &store.Datastore{}
		r := mux.NewRouter()
		cfg := &config.Config{
			RedirectAPIURL:         testRedirectAPIURL,
			EnablePrivateEndpoints: true,
		}
		redirectAPI := api.Setup(context.Background(), r, dataStore, newAuthMiddlwareMock(), cfg)

		Convey("Then all read and write routes should be registered", func() {
			So(hasRoute(redirectAPI.Router, "/v1/redirects/{id}", "GET"), ShouldBeTrue)
			So(hasRoute(redirectAPI.Router, "/v1/redirects", "GET"), ShouldBeTrue)
			So(hasRoute(redirectAPI.Router, "/v1/redirects/{id}", "PUT"), ShouldBeTrue)
			So(hasRoute(redirectAPI.Router, "/v1/redirects/{id}", "DELETE"), ShouldBeTrue)
		})
	})
}

func TestSetupPublicEndpoints(t *testing.T) {
	Convey("Given an API instance with private endpoints disabled", t, func() {
		dataStore := &store.Datastore{}
		r := mux.NewRouter()
		cfg := &config.Config{
			RedirectAPIURL:         testRedirectAPIURL,
			EnablePrivateEndpoints: false,
		}
		redirectAPI := api.Setup(context.Background(), r, dataStore, newAuthMiddlwareMock(), cfg)

		Convey("Then only public GET routes should be registered", func() {
			So(hasRoute(redirectAPI.Router, "/v1/redirects/{id}", "GET"), ShouldBeTrue)
			So(hasRoute(redirectAPI.Router, "/v1/redirects", "GET"), ShouldBeTrue)
		})

		Convey("And write routes should not be registered", func() {
			So(hasRoute(redirectAPI.Router, "/v1/redirects/{id}", "PUT"), ShouldBeFalse)
			So(hasRoute(redirectAPI.Router, "/v1/redirects/{id}", "DELETE"), ShouldBeFalse)
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, http.NoBody)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}

func newAuthMiddlwareMock() *authorisation.MiddlewareMock {
	return &authorisation.MiddlewareMock{
		RequireFunc: func(_ string, handlerFunc http.HandlerFunc) http.HandlerFunc {
			return handlerFunc
		},
	}
}
