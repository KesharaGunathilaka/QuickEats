package httputil

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
)

type PaginationParams struct {
	Cursor string
	Limit  int
}

type PaginationMeta struct {
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"hasMore"`
	Limit   int    `json:"limit"`
}

func ParsePagination(r *http.Request) PaginationParams {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return PaginationParams{
		Cursor: r.URL.Query().Get("cursor"),
		Limit:  limit,
	}
}

func EncodeCursor(data map[string]interface{}) string {
	b, _ := json.Marshal(data)
	return base64.URLEncoding.EncodeToString(b)
}

func DecodeCursor(cursor string) (map[string]interface{}, error) {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return data, nil
}
