package handle

import (
	"log/slog"
	"net/http"
	"npmeta/config"
)

type Middleware struct {
	cfg config.Config
}

func NewMiddleware(cfg config.Config) Middleware {
	return Middleware{cfg: cfg}
}

func (mw Middleware) Wrap(handler func(w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// status, err := handler(w, mw.cfg.ToContext(r))
		// almostrand := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", time.Now().UnixNano())))
		// almostrand = strings.ToUpper(almostrand)
		// almostrand = almostrand[len(almostrand)-10 : len(almostrand)-2]
		status, err := handler(w, r)
		if err == nil {
			return
		}
		if status < 500 {
			slog.Info("request failed", "method", r.Method, "url", r.URL, "http_status", status, "error", err)
		} else {
			slog.Error("error occurred", "method", r.Method, "url", r.URL, "http_status", status, "error", err)
		}
		http.Error(w, http.StatusText(status), status)
	}
}
