package tests

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/random"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:8080"
)

func getAuth(field string) string {

	if err := godotenv.Load("../.env"); err != nil {
		panic("Error loading .env file" + err.Error())
	}

	authPass := os.Getenv("AUTH_PASSWORD")
	authUser := os.Getenv("AUTH_USER")

	if authPass == "" || authUser == "" {
		panic("Authentication credentials are not set in config or environment variables")
	}

	switch field {
	case "password":
		return authPass
	case "user":
		return authUser
	default:
		panic("Invalid field for authentication credentials")
	}

}

func TestURLShortener_HappyPath(t *testing.T) {

	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())

	e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithBasicAuth(getAuth("user"), getAuth("password")).
		Expect().
		Status(201).
		JSON().
		Object().
		ContainsKey("alias")
}

func TestURLShortener_SaveRedirectRemove(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "invalid url",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			//SAVE

			resp := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth(getAuth("user"), getAuth("password"))

			if tc.error != "" {
				resp := resp.Expect().Status(http.StatusBadRequest).
					JSON().
					Object()
				resp.Value("error").String().IsEqual(tc.error)
				resp.NotContainsKey("alias")

				return
			}

			obj := resp.Expect().
				Status(http.StatusCreated).
				JSON().
				Object()

			alias := tc.alias

			if tc.alias != "" {
				obj.Value("alias").String().IsEqual(tc.alias)
			} else {
				obj.Value("alias").String().NotEmpty()

				alias = obj.Value("alias").String().Raw()
			}

			//REDIRECT

			testRedirect(t, alias, tc.url)

			//DELETE

			reqDel := e.DELETE("/"+path.Join("url", alias)).
				WithBasicAuth(getAuth("user"), getAuth("password")).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object()
			reqDel.Value("status").String().IsEqual("OK")

			//REDIRECT AGAIN
			testRedirectNotFound(t, alias)

		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToURL)
}

func testRedirectNotFound(t *testing.T, alias string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	_, err := api.GetRedirect(u.String())
	require.ErrorIs(t, err, api.ErrInvalidStatusCode)
}
