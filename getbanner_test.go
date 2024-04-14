package main

import (
	"avito/Components"
	"avito/Controller"
	"avito/Database"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserBannerHandler(t *testing.T) {
	var response *http.Response
	Database.InitDataBase()
	testServer := httptest.NewServer(http.HandlerFunc(Controller.UserBanner))

	defer testServer.Close()

	url := testServer.URL + "/user_banner?tag_id=1&feature_id=1&use_last_version=false"

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
	}
	request.Header.Set("token", "admin_token")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("want %d, got %d", http.StatusOK, response.StatusCode)
	}
	var banner Components.ShortBanner
	if err = json.NewDecoder(response.Body).Decode(&banner); err != nil {
		t.Error(err)
	}
	expectedBanner := Components.ShortBanner{BannerId: 1, Content: "test banner"}
	if expectedBanner != banner {
		t.Errorf("expectedBanner %v, got %v", expectedBanner, banner)
	}
}
