package service

import (
	"fmt"
	"math/rand"
	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/storage"
	"time"
)

type PRService struct {
	storage *storage.PostgresStorage
}

func NewPRService(storage *storage.PostgresStorage) *PRService {
	return &PRService{storage: storage}
}

func (s *PRService) CreateTeam(team *models.Team) error {
	return s.storage.CreateTeam(team)
}

func (s *PRService) GetTeam(teamName string) (*models.Team, error) {
	return s.storage.GetTeam(teamName)
}

func (s *PRService) SetUserActive(userID string, isActive bool) (*models.User, error) {
	return s.storage.SetUserActive(userID, isActive)
}

func (s *PRService) CreatePR(prID, prName, authorID string) (*models.PullRequest, error) {
	exists, err := s.storage.PRExists(prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("PR_EXISTS")
	}

	authorTeam, err := s.storage.GetUserTeam(authorID)
	if err != nil {
		return nil, fmt.Errorf("NOT_FOUND")
	}

	reviewers, err := s.selectReviewers(authorTeam, authorID)
	if err != nil {
		return nil, err
	}

	pr := &models.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: reviewers,
		CreatedAt:         &[]time.Time{time.Now()}[0],
	}

	if err := s.storage.CreatePR(pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PRService) selectReviewers(teamName, excludeUserID string) ([]string, error) {
	availableReviewers, err := s.storage.GetActiveTeamMembers(teamName, excludeUserID)
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(availableReviewers), func(i, j int) {
		availableReviewers[i], availableReviewers[j] = availableReviewers[j], availableReviewers[i]
	})

	count := min(2, len(availableReviewers))
	if count == 0 {
		return []string{}, nil
	}

	return availableReviewers[:count], nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *PRService) MergePR(prID string) (*models.PullRequest, error) {
	pr, err := s.storage.GetPR(prID)
	if err != nil {
		return nil, err
	}

	if pr.Status == "OPEN" {
		if err := s.storage.MergePR(prID); err != nil {
			return nil, err
		}
		pr, err = s.storage.GetPR(prID)
		if err != nil {
			return nil, err
		}
	}

	return pr, nil
}

func (s *PRService) ReassignReviewer(prID, oldUserID string) (*models.PullRequest, string, error) {
	pr, err := s.storage.GetPR(prID)
	if err != nil {
		return nil, "", err
	}

	if pr.Status == "MERGED" {
		return nil, "", fmt.Errorf("PR_MERGED")
	}

	found := false
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			found = true
			break
		}
	}
	if !found {
		return nil, "", fmt.Errorf("NOT_ASSIGNED")
	}

	reviewerTeam, err := s.storage.GetUserTeam(oldUserID)
	if err != nil {
		return nil, "", fmt.Errorf("NOT_FOUND")
	}

	availableReviewers, err := s.storage.GetActiveTeamMembers(reviewerTeam, oldUserID)
	if err != nil {
		return nil, "", err
	}

	var candidates []string
	for _, candidate := range availableReviewers {
		isAlreadyReviewer := false
		for _, reviewer := range pr.AssignedReviewers {
			if candidate == reviewer {
				isAlreadyReviewer = true
				break
			}
		}
		if !isAlreadyReviewer && candidate != pr.AuthorID {
			candidates = append(candidates, candidate)
		}
	}

	if len(candidates) == 0 {
		return nil, "", fmt.Errorf("NO_CANDIDATE")
	}

	newReviewer := candidates[rand.Intn(len(candidates))]

	newReviewers := make([]string, len(pr.AssignedReviewers))
	for i, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			newReviewers[i] = newReviewer
		} else {
			newReviewers[i] = reviewer
		}
	}

	if err := s.storage.UpdatePRReviewers(prID, newReviewers); err != nil {
		return nil, "", err
	}

	updatedPR, err := s.storage.GetPR(prID)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewer, nil
}

func (s *PRService) GetUserReviewPRs(userID string) ([]models.PullRequestShort, error) {
	return s.storage.GetUserReviewPRs(userID)
}

// GetStats возвращает статистику системы
func (s *PRService) GetStats() (map[string]interface{}, error) {
	return s.storage.GetStats()
}
