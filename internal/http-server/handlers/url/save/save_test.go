package save_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/service"

	"url-shortener/internal/http-server/handlers/url/save"
	mocks "url-shortener/internal/http-server/handlers/url/save/mocks"

	"github.com/stretchr/testify/require"
)

func TestSave_Success(t *testing.T) {
	cases := []struct {
		name  string
		alias string
		url   string
	}{
		{
			name:  "Success",
			alias: "testSuccess",
			url:   "https://google.com",
		},
		{
			name:  "SuccessEmptyAlias",
			alias: "",
			url:   "https://google.com",
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewURLSaver(t)

			urlSaverMock.On("RegisterSaveURL", service.CreateURLInput{
				URL:   tc.url,
				Alias: tc.alias,
			}).
				Return(tc.alias, nil).
				Once()

			handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)
			//assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusCreated, rr.Code)

			body := rr.Body.String()

			var resp save.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, "", resp.Error)

		})
	}
}

func TestSave_BadRequest(t *testing.T) {
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

			urlSaverMock := mocks.NewURLSaver(t)

			if tc.wantMockCall {
				urlSaverMock.On("RegisterSaveURL", service.CreateURLInput{
					URL:   tc.url,
					Alias: tc.alias,
				}).
					Return(tc.alias, tc.mockError).
					Once()
			}

			handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

			input := tc.body
			if input == "" {
				input = fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)
			}

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)
			//assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusBadRequest, rr.Code)

			body := rr.Body.String()

			var resp save.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)

		})
	}
}

func TestSave_Conflict(t *testing.T) {

	url := "https://example.com"
	alias := "testAlreadyExistsAlias"
	urlSaverMock := mocks.NewURLSaver(t)

	urlSaverMock.On("RegisterSaveURL", service.CreateURLInput{
		URL:   url,
		Alias: alias,
	}).
		Return("", service.ErrURLExists).
		Once()

	handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

	input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, url, alias)

	req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)

	body := rr.Body.String()

	var resp save.Response

	require.NoError(t, json.Unmarshal([]byte(body), &resp))

	require.Equal(t, "url already exists", resp.Error)
}
