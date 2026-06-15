package delete_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/delete/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestDelete_Success(t *testing.T) {

	alias := "Success"
	urlDeleterMock := mocks.NewURLDelete(t)

	urlDeleterMock.On("RegisterDeleteURL", alias).
		Return(int64(1), nil).
		Once()

	r := chi.NewRouter()
	r.Delete("/{alias}", delete.New(slogdiscard.NewDiscardLogger(), urlDeleterMock))

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/%s", alias), nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	body := rr.Body.String()

	var resp delete.Response

	require.NoError(t, json.Unmarshal([]byte(body), &resp))

	require.Equal(t, "", resp.Error)
}

func TestDelete_Unavailable(t *testing.T) {

	alias := "testUnavailableAlias"
	urlDeleterMock := mocks.NewURLDelete(t)

	urlDeleterMock.On("RegisterDeleteURL", alias).
		Return(int64(0), service.ErrURLNotFound).
		Once()

	r := chi.NewRouter()
	r.Delete("/{alias}", delete.New(slogdiscard.NewDiscardLogger(), urlDeleterMock))

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/%s", alias), nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	body := rr.Body.String()

	var resp delete.Response

	require.NoError(t, json.Unmarshal([]byte(body), &resp))

	require.Equal(t, "url not found", resp.Error)
}
