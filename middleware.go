package slackactivity

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
)

type (
	SigningSecret string
	Middleware    func(next http.Handler) http.Handler

	SecretsVerifierMiddleware struct {
		signingSecret SigningSecret
	}
)

func (v *SecretsVerifierMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	sv, err := slack.NewSecretsVerifier(r.Header, string(v.signingSecret))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if _, err := sv.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if err := sv.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}
}

func NewSecretsVerifierMiddleware(secret SigningSecret) Middleware {
	secretsVerifier := &SecretsVerifierMiddleware{secret}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secretsVerifier.ServeHTTP(w, r)
			next.ServeHTTP(w, r)
		})
	}
}
