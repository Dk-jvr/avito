package Components

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

func ErrorMaker(writer http.ResponseWriter, err error) {
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		errorMessage := make(map[string]string)
		errorMessage["error"] = err.Error()
		response, _ := json.Marshal(errorMessage)
		writer.Write(response)
	}
}
