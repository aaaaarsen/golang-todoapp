package statistics_transport_http

import (
	"context"
	"net/http"
	"time"

	statistics_service "github.com/aaaaarsen/golang-todoapp/internal/features/statistics/service"
	core_http_server "github.com/aaaaarsen/golang-todoapp/internal/core/transport/http/server"
)

type StatisticsService interface {
	GetStatistics(
		ctx context.Context,
		userID *int,
		from *time.Time,
		to *time.Time,
		category *string,
	) (statistics_service.StatisticsResult, error)
}

type StatisticsHTTPHandler struct {
	statisticsService StatisticsService
}

func NewStatisticsHTTPHandler(statisticsService StatisticsService) *StatisticsHTTPHandler {
	return &StatisticsHTTPHandler{
		statisticsService: statisticsService,
	}
}

func (h *StatisticsHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{Method: http.MethodGet, Path: "/statistics", Handler: h.GetStatistics},
	}
}