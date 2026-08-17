package handler

import (
	"net/http"
	"time"

	"checkinkeeper/internal/model"
	"checkinkeeper/pkg/httpx"
)

func (s *Server) registerActivityRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/activities", s.createActivity)
	mux.HandleFunc("GET /api/activities", s.listActivities)
	mux.HandleFunc("GET /api/activities/{id}", s.getActivity)
	mux.HandleFunc("PUT /api/activities/{id}", s.updateActivity)
	mux.HandleFunc("DELETE /api/activities/{id}", s.deleteActivity)
	mux.HandleFunc("POST /api/activities/{id}/transition", s.transitionActivity)
}

type createActivityRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"` // RFC3339
	EndTime     string `json:"end_time"`   // RFC3339
}

func (s *Server) createActivity(w http.ResponseWriter, r *http.Request) {
	var req createActivityRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	input := model.Activity{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}
	if req.StartTime != "" {
		t, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			httpx.BadRequest(w, "start_time 格式不合法，需为 RFC3339")
			return
		}
		input.StartTime = t
	}
	if req.EndTime != "" {
		t, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			httpx.BadRequest(w, "end_time 格式不合法，需为 RFC3339")
			return
		}
		input.EndTime = t
	}
	a, err := s.svc.CreateActivity(input)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, a)
}

func (s *Server) listActivities(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ActivityFilter{
		Status:  r.URL.Query().Get("status"),
		Keyword: r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListActivities(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getActivity(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetActivity(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

type updateActivityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Server) updateActivity(w http.ResponseWriter, r *http.Request) {
	var req updateActivityRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.UpdateActivity(r.PathValue("id"), req.Name, req.Description)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) deleteActivity(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteActivity(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type transitionRequest struct {
	To string `json:"to"`
}

func (s *Server) transitionActivity(w http.ResponseWriter, r *http.Request) {
	var req transitionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.TransitionActivity(r.PathValue("id"), req.To)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}
