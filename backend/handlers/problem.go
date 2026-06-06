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

type ProblemHandler struct{}

func NewProblemHandler() *ProblemHandler {
	return &ProblemHandler{}
}

type CreateProblemRequest struct {
	Title        string `json:"title"`
	BoardSize    int    `json:"board_size"`
	Goal         string `json:"goal"`
	InitialSGF   string `json:"initial_sgf"`
	SolutionSGF  string `json:"solution_sgf"`
	Description  string `json:"description"`
	Difficulty   string `json:"difficulty"`
	IsPublic     bool   `json:"is_public"`
}

type AttemptRequest struct {
	UserMoves []string `json:"user_moves"`
	TimeSpent int      `json:"time_spent"`
}

func validGoal(g string) bool {
	return g == "black_kill" || g == "black_live" || g == "white_kill" || g == "white_live"
}

func validDifficulty(d string) bool {
	return d == "easy" || d == "medium" || d == "hard" || d == "expert"
}

func (h *ProblemHandler) Create(c echo.Context) error {
	userID := middleware.GetUserID(c)
	var req CreateProblemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if strings.TrimSpace(req.Title) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title is required"})
	}
	if !validGoal(req.Goal) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid goal"})
	}
	if strings.TrimSpace(req.InitialSGF) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "initial_sgf is required"})
	}
	if strings.TrimSpace(req.SolutionSGF) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "solution_sgf is required"})
	}

	if _, err := sgf.Parse(req.InitialSGF); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid initial_sgf: " + err.Error()})
	}
	if _, err := sgf.Parse(req.SolutionSGF); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid solution_sgf: " + err.Error()})
	}

	if req.BoardSize == 0 {
		req.BoardSize = 19
	}
	if req.Difficulty == "" {
		req.Difficulty = "medium"
	}
	if !validDifficulty(req.Difficulty) {
		req.Difficulty = "medium"
	}

	problem := models.Problem{
		UserID:      userID,
		Title:       req.Title,
		BoardSize:   req.BoardSize,
		Goal:        req.Goal,
		InitialSGF:  req.InitialSGF,
		SolutionSGF: req.SolutionSGF,
		Description: req.Description,
		Difficulty:  req.Difficulty,
		IsPublic:    req.IsPublic,
	}

	if err := database.DB.Create(&problem).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create problem"})
	}

	return c.JSON(http.StatusCreated, problem)
}

func (h *ProblemHandler) List(c echo.Context) error {
	var problems []models.Problem
	query := database.DB.Preload("User").Order("created_at DESC")

	userIDParam := c.QueryParam("user_id")
	if userIDParam != "" {
		query = query.Where("user_id = ?", userIDParam)
	} else {
		query = query.Where("is_public = ?", true)
	}

	if goal := c.QueryParam("goal"); goal != "" {
		query = query.Where("goal = ?", goal)
	}
	if diff := c.QueryParam("difficulty"); diff != "" {
		query = query.Where("difficulty = ?", diff)
	}
	if keyword := c.QueryParam("keyword"); keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
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

	if err := query.Find(&problems).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch problems"})
	}

	var total int64
	countQuery := database.DB.Model(&models.Problem{})
	uidParam := c.QueryParam("user_id")
	if uidParam != "" {
		countQuery = countQuery.Where("user_id = ?", uidParam)
	} else {
		countQuery = countQuery.Where("is_public = ?", true)
	}
	if goal := c.QueryParam("goal"); goal != "" {
		countQuery = countQuery.Where("goal = ?", goal)
	}
	if diff := c.QueryParam("difficulty"); diff != "" {
		countQuery = countQuery.Where("difficulty = ?", diff)
	}
	if keyword := c.QueryParam("keyword"); keyword != "" {
		countQuery = countQuery.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	countQuery.Count(&total)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"problems": problems,
		"total":    total,
	})
}

func (h *ProblemHandler) Get(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var problem models.Problem
	if err := database.DB.Preload("User").First(&problem, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "problem not found"})
	}

	return c.JSON(http.StatusOK, problem)
}

func (h *ProblemHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var problem models.Problem
	if err := database.DB.First(&problem, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "problem not found"})
	}

	if problem.UserID != userID && role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "no permission"})
	}

	var req CreateProblemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Title != "" {
		problem.Title = req.Title
	}
	if req.BoardSize > 0 {
		problem.BoardSize = req.BoardSize
	}
	if req.Goal != "" && validGoal(req.Goal) {
		problem.Goal = req.Goal
	}
	if req.InitialSGF != "" {
		problem.InitialSGF = req.InitialSGF
	}
	if req.SolutionSGF != "" {
		problem.SolutionSGF = req.SolutionSGF
	}
	problem.Description = req.Description
	if req.Difficulty != "" && validDifficulty(req.Difficulty) {
		problem.Difficulty = req.Difficulty
	}
	problem.IsPublic = req.IsPublic

	if err := database.DB.Save(&problem).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update problem"})
	}

	return c.JSON(http.StatusOK, problem)
}

func (h *ProblemHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var problem models.Problem
	if err := database.DB.First(&problem, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "problem not found"})
	}

	if problem.UserID != userID && role != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "no permission"})
	}

	if err := database.DB.Delete(&problem).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete problem"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *ProblemHandler) Attempt(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	userID := middleware.GetUserID(c)

	var problem models.Problem
	if err := database.DB.First(&problem, id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "problem not found"})
	}

	var req AttemptRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	isCorrect := false
	solutionTree, err := sgf.Parse(problem.SolutionSGF)
	if err == nil {
		isCorrect = checkSolution(solutionTree, req.UserMoves, problem.Goal)
	}

	userMoves := models.StringSlice(req.UserMoves)
	ts := req.TimeSpent
	attempt := models.ProblemAttempt{
		ProblemID: id,
		UserID:    userID,
		UserMoves: userMoves,
		IsCorrect: isCorrect,
		TimeSpent: &ts,
	}
	database.DB.Create(&attempt)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"is_correct": isCorrect,
		"attempt_id": attempt.ID,
		"solution":   problem.SolutionSGF,
	})
}

func checkSolution(solutionTree *sgf.GameTree, userMoves []string, goal string) bool {
	if len(userMoves) == 0 {
		return false
	}

	var allPaths [][]int
	sgf.CollectAllPaths(solutionTree.Root, []int{0}, &allPaths)

	for _, path := range allPaths {
		if len(path) <= 1 {
			continue
		}
		gs := sgf.NewGameState(solutionTree)
		current := solutionTree.Root
		moves := []string{}
		for i := 1; i < len(path); i++ {
			if path[i] >= len(current.Children) {
				break
			}
			current = current.Children[path[i]]
			if b, ok := current.Properties["B"]; ok && len(b) > 0 && b[0] != "" {
				moves = append(moves, "B:"+b[0])
			}
			if w, ok := current.Properties["W"]; ok && len(w) > 0 && w[0] != "" {
				moves = append(moves, "W:"+w[0])
			}
		}

		if matchMoves(moves, userMoves) {
			return true
		}
		_ = gs
	}

	return false
}

func matchMoves(solution, user []string) bool {
	if len(user) > len(solution) {
		return false
	}
	for i, m := range user {
		if i >= len(solution) {
			return false
		}
		if m != solution[i] {
			return false
		}
	}
	return true
}

func (h *ProblemHandler) ListAttempts(c echo.Context) error {
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid problem id"})
	}
	userID := middleware.GetUserID(c)

	var attempts []models.ProblemAttempt
	query := database.DB.Where("problem_id = ? AND user_id = ?", problemID, userID).Order("created_at DESC")

	if err := query.Find(&attempts).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch attempts"})
	}

	return c.JSON(http.StatusOK, attempts)
}
