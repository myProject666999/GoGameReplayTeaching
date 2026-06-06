package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/goreplay/backend/database"
	"github.com/goreplay/backend/middleware"
	"github.com/goreplay/backend/models"
	"github.com/goreplay/backend/pkg/sgf"
)

type GameHandler struct{}

func NewGameHandler() *GameHandler {
	return &GameHandler{}
}

type CreateGameRequest struct {
	Title        string  `json:"title"`
	BlackPlayer  string  `json:"black_player"`
	WhitePlayer  string  `json:"white_player"`
	BoardSize    int     `json:"board_size"`
	Komi         float64 `json:"komi"`
	Result       string  `json:"result"`
	DatePlayed   string  `json:"date_played"`
	SGFContent   string  `json:"sgf_content"`
	Description  string  `json:"description"`
	IsPublic     bool    `json:"is_public"`
}

func (h *GameHandler) Create(c echo.Context) error {
	userID := middleware.GetUserID(c)
	var req CreateGameRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if strings.TrimSpace(req.SGFContent) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "SGF content is required"})
	}

	if _, err := sgf.Parse(req.SGFContent); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid SGF format: " + err.Error()})
	}

	if req.BoardSize == 0 {
		req.BoardSize = 19
	}
	if req.Title == "" {
		req.Title = "未命名棋谱"
	}

	var datePtr *string
	if req.DatePlayed != "" {
		datePtr = &req.DatePlayed
	}

	game := models.Game{
		UserID:      userID,
		Title:       req.Title,
		BlackPlayer: req.BlackPlayer,
		WhitePlayer: req.WhitePlayer,
		BoardSize:   req.BoardSize,
		Komi:        req.Komi,
		Result:      req.Result,
		DatePlayed:  datePtr,
		SGFContent:  req.SGFContent,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	}

	if err := database.DB.Create(&game).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create game"})
	}

	return c.JSON(http.StatusCreated, game)
}

func (h *GameHandler) List(c echo.Context) error {
	var games []models.Game
	query := database.DB.Preload("User").Order("created_at DESC")

	userIDParam := c.QueryParam("user_id")
	if userIDParam != "" {
		query = query.Where("user_id = ?", userIDParam)
	} else {
		query = query.Where("is_public = ?", true)
	}

	if keyword := c.QueryParam("keyword"); keyword != "" {
		query = query.Where("title LIKE ? OR black_player LIKE ? OR white_player LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if offset := c.QueryParam("offset"); offset != "" {
		if n, err := strconv.Atoi(offset); err == nil && n >= 0 {
			query = query.Offset(n)
		}
	}
	query = query.Limit(limit)

	if err := query.Find(&games).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch games"})
	}

	var total int64
	countQuery := database.DB.Model(&models.Game{})
	if userIDParam != "" {
		countQuery = countQuery.Where("user_id = ?", userIDParam)
	} else {
		countQuery = countQuery.Where("is_public = ?", true)
	}
	countQuery.Count(&total)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"games": games,
		"total": total,
	})
}

func (h *GameHandler) Get(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var game models.Game
	if err := database.DB.Preload("User").First(&game, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "game not found"})
	}

	database.DB.Model(&game).Update("view_count", game.ViewCount+1)

	return c.JSON(http.StatusOK, game)
}

func (h *GameHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var game models.Game
	if err := database.DB.First(&game, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "game not found"})
	}

	if game.UserID != userID && role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "no permission"})
	}

	var req CreateGameRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.SGFContent != "" {
		if _, err := sgf.Parse(req.SGFContent); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid SGF format: " + err.Error()})
		}
		game.SGFContent = req.SGFContent
	}

	if req.Title != "" {
		game.Title = req.Title
	}
	game.BlackPlayer = req.BlackPlayer
	game.WhitePlayer = req.WhitePlayer
	if req.BoardSize > 0 {
		game.BoardSize = req.BoardSize
	}
	game.Komi = req.Komi
	game.Result = req.Result
	game.Description = req.Description
	game.IsPublic = req.IsPublic

	if err := database.DB.Save(&game).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update game"})
	}

	return c.JSON(http.StatusOK, game)
}

func (h *GameHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var game models.Game
	if err := database.DB.First(&game, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "game not found"})
	}

	if game.UserID != userID && role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "no permission"})
	}

	if err := database.DB.Delete(&game).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete game"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *GameHandler) ParseSGF(c echo.Context) error {
	var body struct {
		SGF string `json:"sgf"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	tree, err := sgf.Parse(body.SGF)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	gs := sgf.NewGameState(tree)
	var allPaths [][]int
	sgf.CollectAllPaths(tree.Root, []int{0}, &allPaths)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tree":       tree,
		"board_size": tree.BoardSize,
		"state": map[string]interface{}{
			"board":       gs.Board,
			"move_number": gs.MoveNumber,
			"path":        gs.Path,
		},
		"total_moves": len(allPaths),
		"paths":       allPaths,
	})
}
