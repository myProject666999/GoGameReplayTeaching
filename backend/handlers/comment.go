package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/goreplay/backend/database"
	"github.com/goreplay/backend/middleware"
	"github.com/goreplay/backend/models"
)

type CommentHandler struct{}

func NewCommentHandler() *CommentHandler {
	return &CommentHandler{}
}

type CreateCommentRequest struct {
	GameID       uint64 `json:"game_id"`
	MoveNumber   int    `json:"move_number"`
	NodePath     string `json:"node_path"`
	Content      string `json:"content"`
	VariationSGF string `json:"variation_sgf"`
}

func (h *CommentHandler) Create(c echo.Context) error {
	userID := middleware.GetUserID(c)
	var req CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if req.GameID == 0 || req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "game_id and content are required"})
	}

	comment := models.Comment{
		GameID:       req.GameID,
		MoveNumber:   req.MoveNumber,
		NodePath:     req.NodePath,
		UserID:       userID,
		Content:      req.Content,
		VariationSGF: req.VariationSGF,
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create comment"})
	}

	database.DB.Preload("User").First(&comment, comment.ID)
	return c.JSON(http.StatusCreated, comment)
}

func (h *CommentHandler) ListByGame(c echo.Context) error {
	gameID, err := strconv.ParseUint(c.Param("game_id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid game_id"})
	}

	var comments []models.Comment
	query := database.DB.Preload("User").Where("game_id = ?", gameID).Order("move_number ASC, created_at ASC")

	if moveNum := c.QueryParam("move_number"); moveNum != "" {
		if n, err := strconv.Atoi(moveNum); err == nil {
			query = query.Where("move_number = ?", n)
		}
	}

	if err := query.Find(&comments).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch comments"})
	}

	return c.JSON(http.StatusOK, comments)
}

func (h *CommentHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "comment not found"})
	}

	if comment.UserID != userID && role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "no permission"})
	}

	var req CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Content != "" {
		comment.Content = req.Content
	}
	comment.VariationSGF = req.VariationSGF

	if err := database.DB.Save(&comment).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update comment"})
	}

	return c.JSON(http.StatusOK, comment)
}

func (h *CommentHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "comment not found"})
	}

	if comment.UserID != userID && role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "no permission"})
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete comment"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}
