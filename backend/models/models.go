package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type User struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"`
	Nickname  string    `json:"nickname" gorm:"size:50"`
	Role      string    `json:"role" gorm:"size:20;not null;default:student"`
	Avatar    string    `json:"avatar" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }

type Game struct {
	ID           uint64    `json:"id" gorm:"primaryKey"`
	UserID       uint64    `json:"user_id" gorm:"index;not null"`
	Title        string    `json:"title" gorm:"size:255;not null"`
	BlackPlayer  string    `json:"black_player" gorm:"size:100"`
	WhitePlayer  string    `json:"white_player" gorm:"size:100"`
	BoardSize    int       `json:"board_size" gorm:"not null;default:19"`
	Komi         float64   `json:"komi" gorm:"type:decimal(4,1);default:6.5"`
	Result       string    `json:"result" gorm:"size:50"`
	DatePlayed   *string   `json:"date_played" gorm:"type:date"`
	SGFContent   string    `json:"sgf_content" gorm:"type:longtext;not null"`
	Description  string    `json:"description" gorm:"type:text"`
	IsPublic     bool      `json:"is_public" gorm:"not null;default:false"`
	ViewCount    int       `json:"view_count" gorm:"not null;default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	User         *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (Game) TableName() string { return "games" }

type MoveMarker struct {
	ID         uint64    `json:"id" gorm:"primaryKey"`
	GameID     uint64    `json:"game_id" gorm:"index:idx_game_move;not null"`
	MoveNumber int       `json:"move_number" gorm:"index:idx_game_move;not null"`
	NodePath   string    `json:"node_path" gorm:"size:255"`
	MarkerType string    `json:"marker_type" gorm:"size:20;not null"`
	UserID     *uint64   `json:"user_id" gorm:"index"`
	Note       string    `json:"note" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (MoveMarker) TableName() string { return "move_markers" }

type Comment struct {
	ID           uint64    `json:"id" gorm:"primaryKey"`
	GameID       uint64    `json:"game_id" gorm:"index:idx_game_move;not null"`
	MoveNumber   int       `json:"move_number" gorm:"index:idx_game_move;not null"`
	NodePath     string    `json:"node_path" gorm:"size:255"`
	UserID       uint64    `json:"user_id" gorm:"index;not null"`
	Content      string    `json:"content" gorm:"type:text;not null"`
	VariationSGF string    `json:"variation_sgf" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	User         *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (Comment) TableName() string { return "comments" }

type Problem struct {
	ID          uint64    `json:"id" gorm:"primaryKey"`
	UserID      uint64    `json:"user_id" gorm:"index;not null"`
	Title       string    `json:"title" gorm:"size:255;not null"`
	BoardSize   int       `json:"board_size" gorm:"not null;default:19"`
	Goal        string    `json:"goal" gorm:"size:20;not null"`
	InitialSGF  string    `json:"initial_sgf" gorm:"type:text;not null"`
	SolutionSGF string    `json:"solution_sgf" gorm:"type:text;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Difficulty  string    `json:"difficulty" gorm:"size:20;not null;default:medium"`
	IsPublic    bool      `json:"is_public" gorm:"not null;default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (Problem) TableName() string { return "problems" }

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan StringSlice")
	}
	return json.Unmarshal(bytes, s)
}

type ProblemAttempt struct {
	ID         uint64      `json:"id" gorm:"primaryKey"`
	ProblemID  uint64      `json:"problem_id" gorm:"index:idx_problem_user;not null"`
	UserID     uint64      `json:"user_id" gorm:"index:idx_problem_user;index;not null"`
	UserMoves  StringSlice `json:"user_moves" gorm:"type:text"`
	IsCorrect  bool        `json:"is_correct" gorm:"not null;default:false"`
	TimeSpent  *int        `json:"time_spent"`
	CreatedAt  time.Time   `json:"created_at"`
}

func (ProblemAttempt) TableName() string { return "problem_attempts" }

type GameFavorite struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	UserID    uint64    `json:"user_id" gorm:"uniqueIndex:uk_user_game;not null"`
	GameID    uint64    `json:"game_id" gorm:"uniqueIndex:uk_user_game;not null"`
	CreatedAt time.Time `json:"created_at"`
}

func (GameFavorite) TableName() string { return "game_favorites" }
