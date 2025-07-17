package steps

import "github.com/maxcnunes/httpfake"

// FakeAPI contains all the information for a fake component API
type FakeAPI struct {
	fakeHTTP *httpfake.HTTPFake
}

// NewFakeAPI creates a new fake component API
func NewFakeAPI() *FakeAPI {
	return &FakeAPI{
		fakeHTTP: httpfake.New(),
	}
}

// Close closes the fake API
func (f *FakeAPI) Close() {
	f.fakeHTTP.Close()
}

// Stub /identity endpoint to simulate valid service identity
func (f *FakeAPI) setJSONResponseForGetIdentity() {
	f.fakeHTTP.NewHandler().
		Get("/identity").
		Reply(200).
		BodyString(`{ "identifier": "dis-redirect-api", "token_type": "Service" }`)
}

// Stub /v1/permissions-bundle to return valid permission for dis-redirect-api
func (f *FakeAPI) setJSONResponseForPermissionsBundle() {
	f.fakeHTTP.NewHandler().
		Get("/v1/permissions-bundle").
		Reply(200).
		BodyString(`{
			"legacy:edit": {
				"users/dis-redirect-api": [
					{ "id": "1" }
				],
				"groups/role-admin": [
					{ "id": "1" }
				]
			}
		}`)
}

// Call both stubs in test setup
func (f *FakeAPI) setupDefaultAuthResponses() {
	f.setJSONResponseForGetIdentity()
	f.setJSONResponseForPermissionsBundle()
}
