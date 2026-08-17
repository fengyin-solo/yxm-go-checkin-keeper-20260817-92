package handler

import (
	"net/http"

	"checkinkeeper/internal/model"
	"checkinkeeper/pkg/httpx"
)

func (s *Server) registerRewardRuleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/reward-rules", s.createRewardRule)
	mux.HandleFunc("GET /api/reward-rules", s.listRewardRules)
	mux.HandleFunc("GET /api/reward-rules/{id}", s.getRewardRule)
	mux.HandleFunc("PUT /api/reward-rules/{id}", s.updateRewardRule)
	mux.HandleFunc("DELETE /api/reward-rules/{id}", s.deleteRewardRule)
}

type createRewardRuleRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	ActivityID string `json:"activity_id"`
	Threshold  int    `json:"threshold"`
	Points     int64  `json:"points"`
}

func (s *Server) createRewardRule(w http.ResponseWriter, r *http.Request) {
	var req createRewardRuleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	rule, err := s.svc.CreateRewardRule(model.RewardRule{
		Name:       req.Name,
		Type:       req.Type,
		ActivityID: req.ActivityID,
		Threshold:  req.Threshold,
		Points:     req.Points,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, rule)
}

func (s *Server) listRewardRules(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.RewardRuleFilter{
		Type:       r.URL.Query().Get("type"),
		Status:     r.URL.Query().Get("status"),
		ActivityID: r.URL.Query().Get("activity_id"),
	}
	items, total, err := s.svc.ListRewardRules(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getRewardRule(w http.ResponseWriter, r *http.Request) {
	rule, err := s.svc.GetRewardRule(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rule)
}

type updateRewardRuleRequest struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Threshold int    `json:"threshold"`
	Points    int64  `json:"points"`
}

func (s *Server) updateRewardRule(w http.ResponseWriter, r *http.Request) {
	var req updateRewardRuleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	rule, err := s.svc.UpdateRewardRule(r.PathValue("id"), req.Name, req.Status, req.Threshold, req.Points)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rule)
}

func (s *Server) deleteRewardRule(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteRewardRule(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
