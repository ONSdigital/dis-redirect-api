package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-redirect-api/store"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		redirectStore := &store.Datastore{}
		redirectAPI := GetRedirectAPIWithMocks(*redirectStore)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(redirectAPI.Router, "/v1/redirects/{id}", "GET"), ShouldBeTrue)
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, http.NoBody)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
