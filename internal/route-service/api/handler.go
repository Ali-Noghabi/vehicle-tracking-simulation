package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"vehicle-tracking-simulation/internal/route-service/models"
	"vehicle-tracking-simulation/internal/route-service/service"
)

// Handler handles HTTP requests for the route service
type Handler struct {
	routeFinder *service.RouteFinder
	router      *mux.Router
}

// NewHandler creates a new API handler
func NewHandler(routeFinder *service.RouteFinder) *Handler {
	h := &Handler{
		routeFinder: routeFinder,
		router:      mux.NewRouter(),
	}
	h.setupRoutes()
	return h
}

// setupRoutes configures the API routes
func (h *Handler) setupRoutes() {
	h.router.HandleFunc("/health", h.HealthCheck).Methods("GET")
	h.router.HandleFunc("/api/v1/route", h.FindRoute).Methods("POST")
	h.router.HandleFunc("/api/v1/route/waypoints", h.FindRouteWithWaypoints).Methods("POST")
	h.router.HandleFunc("/api/v1/provider", h.GetProvider).Methods("GET")
}

// GetRouter returns the configured router
func (h *Handler) GetRouter() *mux.Router {
	return h.router
}

// HealthCheck returns the health status of the service
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "healthy",
		"service":  "route-service",
		"provider": h.routeFinder.GetProvider().ProviderName(),
	})
}

// FindRoute handles POST /api/v1/route
// Request body: {"start": {"latitude": 51.5, "longitude": -0.1}, "end": {"latitude": 51.51, "longitude": -0.12}}
func (h *Handler) FindRoute(w http.ResponseWriter, r *http.Request) {
	var req models.RouteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	routeResp, err := h.routeFinder.FindRoute(req)
	if err != nil {
		log.Printf("Error finding route: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, routeResp)
}

// FindRouteWithWaypoints handles POST /api/v1/route/waypoints
// Request body: {"waypoints": [{"latitude": 51.5, "longitude": -0.1}, ...], "profile": "car"}
type WaypointsRequest struct {
	Waypoints []models.Coordinate `json:"waypoints"`
	Profile   string              `json:"profile"`
}

func (h *Handler) FindRouteWithWaypoints(w http.ResponseWriter, r *http.Request) {
	var req WaypointsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if len(req.Waypoints) < 2 {
		respondError(w, http.StatusBadRequest, "At least 2 waypoints required")
		return
	}

	routeResp, err := h.routeFinder.FindRouteWithWaypoints(req.Waypoints, req.Profile)
	if err != nil {
		log.Printf("Error finding route with waypoints: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, routeResp)
}

// GetProvider returns information about the current routing provider
func (h *Handler) GetProvider(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"provider": h.routeFinder.GetProvider().ProviderName(),
	})
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, map[string]interface{}{
		"error": message,
	})
}
