package handler

import (
	"net/http"
	"time"

	"checkinkeeper/internal/model"
	"checkinkeeper/pkg/httpx"
)

func (s *Server) registerCheckinRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/checkins", s.checkin)
	mux.HandleFunc("GET /api/checkins", s.listCheckins)
	mux.HandleFunc("GET /api/checkins/{id}", s.getCheckin)
	mux.HandleFunc("GET /api/users/{id}/streak", s.getUserStreak)
	mux.HandleFunc("GET /api/users/{id}/calendar", s.getUserCalendar)
}

type checkinRequest struct {
	ActivityID string `json:"activity_id"`
	UserID     string `json:"user_id"`
}

func (s *Server) checkin(w http.ResponseWriter, r *http.Request) {
	var req checkinRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	record, err := s.svc.Checkin(req.ActivityID, req.UserID, time.Now())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, record)
}

func (s *Server) listCheckins(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.CheckinFilter{
		ActivityID: r.URL.Query().Get("activity_id"),
		UserID:     r.URL.Query().Get("user_id"),
		FromDate:   r.URL.Query().Get("from_date"),
		ToDate:     r.URL.Query().Get("to_date"),
	}
	items, total, err := s.svc.ListCheckinRecords(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getCheckin(w http.ResponseWriter, r *http.Request) {
	record, err := s.svc.GetCheckinRecord(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, record)
}

func (s *Server) getUserStreak(w http.ResponseWriter, r *http.Request) {
	activityID := r.URL.Query().Get("activity_id")
	if activityID == "" {
		httpx.BadRequest(w, "activity_id 不能为空")
		return
	}
	streak, err := s.svc.GetUserStreak(r.PathValue("id"), activityID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, streak)
}

func (s *Server) getUserCalendar(w http.ResponseWriter, r *http.Request) {
	activityID := r.URL.Query().Get("activity_id")
	month := r.URL.Query().Get("month")
	if activityID == "" {
		httpx.BadRequest(w, "activity_id 不能为空")
		return
	}
	if month == "" {
		httpx.BadRequest(w, "month 不能为空")
		return
	}
	calendar, err := s.svc.GetUserCalendar(r.PathValue("id"), activityID, month)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, calendar)
}
