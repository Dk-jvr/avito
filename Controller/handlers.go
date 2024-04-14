package Controller

import (
	"avito/Components"
	"avito/Database"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"net/http"
	"reflect"
	"strconv"
)

func UserBanner(writer http.ResponseWriter, request *http.Request) {
	var (
		err              error
		tagId, featureId int64
		useLastVersion   bool
		banner           Components.ShortBanner
	)
	defer func() {
		Components.ErrorMaker(writer, err)
	}()
	tagId, err = strconv.ParseInt(request.URL.Query().Get("tag_id"), 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	featureId, err = strconv.ParseInt(request.URL.Query().Get("feature_id"), 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	useLastVersion, err = strconv.ParseBool(request.URL.Query().Get("use_last_version"))
	if err != nil {
		useLastVersion = false
	}
	token := request.Header.Get("token")
	banner, err = Database.GetUserBanner(tagId, featureId, useLastVersion, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writer.WriteHeader(http.StatusNotFound)
			return
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	response, _ := json.Marshal(banner)
	writer.Write(response)
}

func BannerProcessing(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		var (
			tagId, featureId, limit, offset int64
			banners                         []Components.Banner
			response                        []byte
			err                             error
		)

		defer func() {
			Components.ErrorMaker(writer, err)
		}()

		tagId, _ = strconv.ParseInt(request.URL.Query().Get("tag_id"), 10, 64)
		featureId, _ = strconv.ParseInt(request.URL.Query().Get("feature_id"), 10, 64)
		limit, _ = strconv.ParseInt(request.URL.Query().Get("limit"), 10, 64)
		offset, _ = strconv.ParseInt(request.URL.Query().Get("offset"), 10, 64)

		banners, err = Database.GetBanners(tagId, featureId, limit, offset, request.Header.Get("token"))
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(banners) == 0 {
			response, _ = json.Marshal("{}")
		} else {
			response, _ = json.Marshal(banners)
		}
		writer.Write(response)
		return

	case http.MethodPost:
		banner := new(Components.Banner)
		var (
			err error
		)
		defer func() {
			Components.ErrorMaker(writer, err)
		}()

		err = json.NewDecoder(request.Body).Decode(banner)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		validate := validator.New()
		err = validate.Struct(banner)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		err = Database.CreateBanner(*banner)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusCreated)
		return
	}
}

func AdminBannerProcessing(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPatch:
		vars := mux.Vars(request)
		bannerIdString := vars["id"]
		bannerId, err := strconv.ParseInt(bannerIdString, 10, 64)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		defer func() {
			Components.ErrorMaker(writer, err)
		}()
		bannerData := make(map[string]interface{})
		err = json.NewDecoder(request.Body).Decode(&bannerData)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		for key, value := range bannerData {
			switch key {
			case "tag_ids":
				tagIdsFloat, ok := value.([]interface{})
				if !ok {
					writer.WriteHeader(http.StatusBadRequest)
					err = errors.New(fmt.Sprintf("Wrong type of field %s", "tag_ids"))
					return
				}
				tagIds := make([]int, len(tagIdsFloat))
				for i, v := range tagIdsFloat {
					var tagId float64
					tagId, ok = v.(float64)
					if !ok {
						writer.WriteHeader(http.StatusBadRequest)
						err = errors.New(fmt.Sprintf("Wrong type type in field %s, found %s, expected int", "tag_ids", reflect.TypeOf(v)))
						return
					}
					tagIds[i] = int(tagId)
				}
				bannerData["tag_ids"] = pq.Array(tagIds)
			case "feature_id":
				featureId, ok := value.(float64)
				if !ok {
					writer.WriteHeader(http.StatusBadRequest)
					err = errors.New(fmt.Sprintf("Wrong type type in field %s, found %s, expected int", "feature_id", reflect.TypeOf(featureId)))
					return
				}
				bannerData["feature_id"] = int(featureId)
			case "content":
				if content, ok := value.(string); !ok {
					writer.WriteHeader(http.StatusBadRequest)
					err = errors.New(fmt.Sprintf("Wrong type type in field %s, found %s, expected string", "feature_id", reflect.TypeOf(content)))
					return
				}
			case "is_active":
				if isActive, ok := value.(bool); !ok {
					writer.WriteHeader(http.StatusBadRequest)
					err = errors.New(fmt.Sprintf("Wrong type type in field %s, found %s, expected int", "feature_id", reflect.TypeOf(isActive)))
					return
				}
			}
		}
		err = Database.UpdateBanner(bannerId, bannerData)
		if errors.Is(err, sql.ErrNoRows) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		return

	case http.MethodDelete:
		vars := mux.Vars(request)
		bannerIdString := vars["id"]
		bannerId, err := strconv.ParseInt(bannerIdString, 10, 64)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		err = Database.DeleteBanner(bannerId)
		if errors.Is(err, sql.ErrNoRows) {
			writer.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	case http.MethodGet:
		vars := mux.Vars(request)
		bannerIdString := vars["id"]
		var (
			response []byte
			banners  []Components.Banner
		)
		bannerId, err := strconv.ParseInt(bannerIdString, 10, 64)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		banners, err = Database.GetOldBanners(bannerId)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(banners) == 0 {
			response, _ = json.Marshal("{}")
		} else {
			response, _ = json.Marshal(banners)
		}
		writer.Write(response)
		return
	}
}

func ReturningOldBanner(writer http.ResponseWriter, request *http.Request) {
	var version int64
	vars := mux.Vars(request)
	bannerIdString := vars["id"]
	bannerId, err := strconv.ParseInt(bannerIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	version, err = strconv.ParseInt(vars["version"], 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	err = Database.ReturnOldBanner(version, bannerId)
	if errors.Is(err, sql.ErrNoRows) {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func UserAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		token := request.Header.Get("token")
		if token != "admin_token" && token != "user_token" {
			http.Error(writer, "Пользователь не авторизован", http.StatusUnauthorized)
			return
		}
		if request.URL.Path != "/user_banner" && token == "user_token" {
			http.Error(writer, "Пользователь не имеет доступа", http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
