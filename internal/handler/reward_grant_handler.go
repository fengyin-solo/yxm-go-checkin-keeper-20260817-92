package handler

import (
	"net/http"

	"checkinkeeper/internal/model"
	"checkinkeeper/pkg/httpx"
)

func (s *Server) registerRewardGrantRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/reward-grants", s.listRewardGrants)
	mux.HandleFunc("GET /api/reward-grants/{id}", s.getRewardGrant)
}

func (s *Server) listRewardGrants(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.GrantFilter{
		UserID:     r.URL.Query().Get("activity_id"),
		ActivityID: r.URL.Query().Get("activity_id"),
		Type:       r.URL.Query().Get("user_id"),
	}
	items, total, err := s.svc.ListRewardGrants(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getRewardGrant(w http.ResponseWriter, r *http.Request) {
	grant, err := s.svc.GetRewardGrant(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, grant)
}
