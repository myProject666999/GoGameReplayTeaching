package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/goreplay/backend/database"
	"github.com/goreplay/backend/middleware"
	"github.com/goreplay/backend/models"
)

type MarkerHandler struct{}

func NewMarkerHandler() *MarkerHandler {
	return &MarkerHandler{}
}

type CreateMarkerRequest struct {
	GameID     uint64 `json:"game_id"`
	MoveNumber int    `json:"move_number"`
	NodePath   string `json:"node_path"`
	MarkerType string `json:"marker_type"`
	Note       string `json:"note"`
}

func (h *MarkerHandler) Create(c echo.Context) error {
	userID := middleware.GetUserID(c)
	var req CreateMarkerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if req.GameID == 0 || req.MarkerType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "game_id and marker_type are required"})
	}

	validTypes := map[string]bool{
		"black_adv": true, "white_adv": true, "key": true, "question": true, "good": true,
	}
	if !validTypes[req.MarkerType] {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid marker_type"})
	}

	uid := userID
	marker := models.MoveMarker{
		GameID:     req.GameID,
		MoveNumber: req.MoveNumber,
		NodePath:   req.NodePath,
		MarkerType: req.MarkerType,
		UserID:     &uid,
		Note:       req.Note,
	}

	if err := database.DB.Create(&marker).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create marker"})
	}

	return c.JSON(http.StatusCreated, marker)
}

func (h *MarkerHandler) ListByGame(c echo.Context) error {
	gameID, err := strconv.ParseUint(c.Param("game_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid game_id"})
	}

	var markers []models.MoveMarker
	query := database.DB.Preload("User").Where("game_id = ?", gameID).Order("move_number ASC")

	if mt := c.QueryParam("marker_type"); mt != "" {
		query = query.Where("marker_type = ?", mt)
	}

	if err := query.Find(&markers).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch markers"})
	}

	return c.JSON(http.StatusOK, markers)
}

func (h *MarkerHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var marker models.MoveMarker
	if err := database.DB.First(&marker, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "marker not found"})
	}

	if marker.UserID != nil && *marker.UserID != userID && role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "no permission"})
	}

	if err := database.DB.Delete(&marker).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete marker"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}
