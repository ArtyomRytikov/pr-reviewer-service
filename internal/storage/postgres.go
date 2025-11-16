package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"pr-reviewer-service/internal/models"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *PostgresStorage) CreateTeam(team *models.Team) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", team.TeamName).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("TEAM_EXISTS")
	}

	_, err = tx.Exec("INSERT INTO teams (team_name) VALUES ($1)", team.TeamName)
	if err != nil {
		return err
	}

	for _, member := range team.Members {
		_, err = tx.Exec(`
			INSERT INTO users (user_id, username, team_name, is_active) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO UPDATE SET 
				username = EXCLUDED.username,
				team_name = EXCLUDED.team_name,
				is_active = EXCLUDED.is_active
		`, member.UserID, member.Username, team.TeamName, member.IsActive)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStorage) GetTeam(teamName string) (*models.Team, error) {
	var team models.Team
	team.TeamName = teamName

	rows, err := s.db.Query(`
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name = $1
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		team.Members = append(team.Members, member)
	}

	if len(team.Members) == 0 {
		return nil, fmt.Errorf("NOT_FOUND")
	}

	return &team, nil
}

func (s *PostgresStorage) SetUserActive(userID string, isActive bool) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		UPDATE users SET is_active = $1 
		WHERE user_id = $2 
		RETURNING user_id, username, team_name, is_active
	`, isActive, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("NOT_FOUND")
	}

	return &user, err
}

func (s *PostgresStorage) CreatePR(pr *models.PullRequest) error {
	reviewersJSON, _ := json.Marshal(pr.AssignedReviewers)

	_, err := s.db.Exec(`
		INSERT INTO pull_requests 
		(pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, reviewersJSON, time.Now())

	return err
}

func (s *PostgresStorage) GetPR(prID string) (*models.PullRequest, error) {
	var pr models.PullRequest
	var reviewersJSON string
	var mergedAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT pull_request_id, pull_request_name, author_id, status, 
		       assigned_reviewers, created_at, merged_at
		FROM pull_requests 
		WHERE pull_request_id = $1
	`, prID).Scan(
		&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status,
		&reviewersJSON, &pr.CreatedAt, &mergedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}

	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	json.Unmarshal([]byte(reviewersJSON), &pr.AssignedReviewers)
	return &pr, nil
}

func (s *PostgresStorage) MergePR(prID string) error {
	_, err := s.db.Exec(`
		UPDATE pull_requests 
		SET status = 'MERGED', merged_at = $1 
		WHERE pull_request_id = $2 AND status = 'OPEN'
	`, time.Now(), prID)
	return err
}

func (s *PostgresStorage) UpdatePRReviewers(prID string, reviewers []string) error {
	reviewersJSON, _ := json.Marshal(reviewers)
	_, err := s.db.Exec(`
		UPDATE pull_requests 
		SET assigned_reviewers = $1 
		WHERE pull_request_id = $2
	`, reviewersJSON, prID)
	return err
}

func (s *PostgresStorage) GetUserReviewPRs(userID string) ([]models.PullRequestShort, error) {
	rows, err := s.db.Query(`
		SELECT pull_request_id, pull_request_name, author_id, status
		FROM pull_requests 
		WHERE $1 = ANY(assigned_reviewers)
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

func (s *PostgresStorage) GetActiveTeamMembers(teamName string, excludeUserID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT user_id 
		FROM users 
		WHERE team_name = $1 AND is_active = true AND user_id != $2
	`, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (s *PostgresStorage) PRExists(prID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)
	`, prID).Scan(&exists)
	return exists, err
}

func (s *PostgresStorage) UserExists(userID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)
	`, userID).Scan(&exists)
	return exists, err
}

func (s *PostgresStorage) GetUserTeam(userID string) (string, error) {
	var teamName string
	err := s.db.QueryRow(`
		SELECT team_name FROM users WHERE user_id = $1
	`, userID).Scan(&teamName)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("NOT_FOUND")
	}
	return teamName, err
}

// GetStats возвращает статистику системы
func (s *PostgresStorage) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Общая статистика
	var totalTeams, totalUsers, totalPRs, openPRs, mergedPRs int
	err := s.db.QueryRow("SELECT COUNT(*) FROM teams").Scan(&totalTeams)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow("SELECT COUNT(*) FROM pull_requests").Scan(&totalPRs)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow("SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN'").Scan(&openPRs)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow("SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED'").Scan(&mergedPRs)
	if err != nil {
		return nil, err
	}

	// Статистика по назначениям ревьюверов - исправленный запрос
	rows, err := s.db.Query(`
		SELECT value as user_id, COUNT(*) as assignment_count 
		FROM pull_requests, jsonb_array_elements_text(assigned_reviewers) 
		GROUP BY value 
		ORDER BY assignment_count DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topReviewers := make([]map[string]interface{}, 0)
	for rows.Next() {
		var userID string
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, err
		}
		topReviewers = append(topReviewers, map[string]interface{}{
			"user_id":          userID,
			"assignment_count": count,
		})
	}

	stats["total_teams"] = totalTeams
	stats["total_users"] = totalUsers
	stats["total_prs"] = totalPRs
	stats["open_prs"] = openPRs
	stats["merged_prs"] = mergedPRs
	stats["top_reviewers"] = topReviewers

	return stats, nil
}
