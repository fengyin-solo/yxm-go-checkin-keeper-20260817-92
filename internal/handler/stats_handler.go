package handler

import (
	"net/http"
	"strconv"

	"checkinkeeper/pkg/httpx"
)

func (s *Server) registerStatsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stats/overview", s.statsOverview)
	mux.HandleFunc("GET /api/stats/by-activity", s.statsByActivity)
	mux.HandleFunc("GET /api/stats/by-day", s.statsByDay)
	mux.HandleFunc("GET /api/stats/leaderboard", s.statsLeaderboard)
}

func (s *Server) statsOverview(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.Stats())
}

func (s *Server) statsByActivity(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.StatsByActivity())
}

func (s *Server) statsByDay(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.StatsByDay())
}

func (s *Server) statsLeaderboard(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	httpx.OK(w, s.svc.Leaderboard(n))
}
