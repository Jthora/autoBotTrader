package httpapi

import "net/http"

func NewRouter(h *Handlers) *http.ServeMux {
    mux := http.NewServeMux()
    mux.HandleFunc("/health", h.Health)
    mux.HandleFunc("/astrology", h.Astrology)
    mux.HandleFunc("/gravimetrics", h.Gravimetrics)
    mux.HandleFunc("/predict", h.Predict)
    mux.HandleFunc("/push", h.Push)
    return mux
}
