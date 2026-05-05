package statistics_transport_http

import (
	"net/http"
	"strconv"
	"time"

	core_logger "github.com/aaaaarsen/golang-todoapp/internal/core/logger"
	core_http_response "github.com/aaaaarsen/golang-todoapp/internal/core/transport/http/response"
	"go.uber.org/zap"
)

func (h *StatisticsHTTPHandler) GetStatistics(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	log.Debug("invoke GetStatistics handler")

	q := r.URL.Query()

	var userID *int
	if v := q.Get("user_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			http.Error(rw, "invalid user_id", http.StatusBadRequest)
			return
		}
		userID = &id
	}

	var from *time.Time
	if v := q.Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			http.Error(rw, "invalid from format, use RFC3339", http.StatusBadRequest)
			return
		}
		utc := t.UTC()
		from = &utc
	}

	var to *time.Time
	if v := q.Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			http.Error(rw, "invalid to format, use RFC3339", http.StatusBadRequest)
			return
		}
		utc := t.UTC()
		to = &utc
	}

	var category *string
	if v := q.Get("category"); v != "" {
		category = &v
	}

	result, err := h.statisticsService.GetStatistics(ctx, userID, from, to, category)
	if err != nil {
		log.Error("get statistics error", zap.Error(err))
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}

	responseHandler.JSONResponse(&result, http.StatusOK)
}