package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-redirect-api/store"
	authorisation "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		dataStore := &store.Datastore{}
		redirectAPI := GetRedirectAPIWithMocks(*dataStore)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(redirectAPI.Router, "/v1/redirects/{id}", "PUT"), ShouldBeTrue)
			So(hasRoute(redirectAPI.Router, "/v1/redirects/{id}", "GET"), ShouldBeTrue)
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
		RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
			return handlerFunc
		},
	}
}
