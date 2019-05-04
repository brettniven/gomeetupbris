package gomeetupbris

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

// NewHandler creates the http handler
func NewHandler() http.Handler {

	r := mux.NewRouter()

	r.Path("/ping").
		Methods(http.MethodGet).
		HandlerFunc(pingHandler())

	// TODO health/readiness K8's endpoints

	return r
}

func pingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ping_ok"))
		if err != nil {
			logrus.WithError(err).Error()
		}
	}
}
