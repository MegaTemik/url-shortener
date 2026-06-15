package update_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/url/update"
	"url-shortener/internal/http-server/handlers/url/update/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/service"

	"github.com/stretchr/testify/require"
)

func TestUpdate_Success(t *testing.T) {

	alias := "testSuccess"
	url := "https://google.com"

	urlUpdaterMock := mocks.NewURLUpdater(t)

	urlUpdaterMock.On("RegisterUpdateURL", service.UpdateURLInput{
		Alias:  alias,
		NewURL: url,
	}).
		Return(int64(1), nil).
		Once()

	handler := update.New(slogdiscard.NewDiscardLogger(), urlUpdaterMock)

	input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, url, alias)

	req, err := http.NewRequest(http.MethodPost, "/update", bytes.NewReader([]byte(input)))
	require.NoError(t, err)
	//assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	body := rr.Body.String()

	var resp update.Response

	require.NoError(t, json.Unmarshal([]byte(body), &resp))

	require.Equal(t, "", resp.Error)
}

func TestUpdate_BadRequest(t *testing.T) {
	cases := []struct {
		name         string
		alias        string
		url          string
		body         string
		respError    string
		wantMockCall bool
		mockError    error
	}{
		{
			name:         "Empty URL",
			alias:        "testAlias",
			url:          "",
			body:         "",
			respError:    "invalid url",
			wantMockCall: true,
			mockError:    service.ErrInvalidURL,
		},
		{
			name:         "Invalid URL",
			alias:        "testInvalidAlias",
			url:          "some invalid URL",
			body:         "",
			respError:    "invalid url",
			wantMockCall: true,
			mockError:    service.ErrInvalidURL,
		},
		{
			name:         "Bad JSON",
			alias:        "testBadJSONAlias",
			url:          "https://example.com",
			body:         `{"url": "https://example.com", "json":`,
			respError:    "failed to decode request",
			wantMockCall: false,
			mockError:    nil,
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlUpdaterMock := mocks.NewURLUpdater(t)

			if tc.wantMockCall {
				urlUpdaterMock.On("RegisterUpdateURL", service.UpdateURLInput{
					Alias:  tc.alias,
					NewURL: tc.url,
				}).
					Return(int64(0), tc.mockError).
					Once()
			}

			handler := update.New(slogdiscard.NewDiscardLogger(), urlUpdaterMock)

			input := tc.body
			if input == "" {
				input = fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)
			}

			req, err := http.NewRequest(http.MethodPost, "/update", bytes.NewReader([]byte(input)))
			require.NoError(t, err)
			//assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusBadRequest, rr.Code)

			body := rr.Body.String()

			var resp update.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)

		})
	}
}

func TestUpdate_StatusNotFound(t *testing.T) {

	alias := ""
	url := "https://google.com"

	urlUpdaterMock := mocks.NewURLUpdater(t)

	urlUpdaterMock.On("RegisterUpdateURL", service.UpdateURLInput{
		Alias:  alias,
		NewURL: url,
	}).
		Return(int64(0), service.ErrURLNotFound).
		Once()

	handler := update.New(slogdiscard.NewDiscardLogger(), urlUpdaterMock)

	input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, url, alias)

	req, err := http.NewRequest(http.MethodPost, "/update", bytes.NewReader([]byte(input)))
	require.NoError(t, err)
	//assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	body := rr.Body.String()

	var resp update.Response

	require.NoError(t, json.Unmarshal([]byte(body), &resp))

	require.Equal(t, "url not found", resp.Error)
}
