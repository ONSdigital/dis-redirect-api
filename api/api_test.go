package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-redirect-api/api"
	apimock "github.com/ONSdigital/dis-redirect-api/api/mock"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		r := mux.NewRouter()
		ctx := context.Background()
		mockRedis := &apimock.RedisClientMock{}

		redirectAPI := api.Setup(ctx, r, mockRedis)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(redirectAPI.Router, "/redirects/{id}", "GET"), ShouldBeTrue)
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, http.NoBody)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}

func setupAPI(redirect api.RedisClient) *api.RedirectAPI {
	return api.Setup(context.Background(), mux.NewRouter(), redirect)
}
