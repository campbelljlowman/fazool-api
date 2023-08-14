package session

import (
	"time"

	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
)

type sessionMetrics struct {
	gorm.Model
	StartedAt 				time.Time
	EndedAt 				time.Time
	SessionID				int
	AdminAccountID			int
	NumberOfVotes			int
	NumberOfVoters			int
	NumberOfSongsPlayed		int
	NumberOfBonusVotesAdded	int
	NumberOfSuperVoters		int
	NumberOfBonusVoters		int
	superVoterMap			map[string]struct{} `gorm:"-:all"`
	bonusVoterMap			map[int]struct{} `gorm:"-:all"`
}

func (s *SessionServiceInMemory) GetActiveSessionMetrics(c *gin.Context) {
	var activeSessionMetrics [] sessionMetrics
	s.allSessionsMutex.Lock()
	for _, session := range(s.sessions) {
		activeSessionMetrics = append(activeSessionMetrics, *session.sessionMetrics)
	}
	s.allSessionsMutex.Unlock()

	c.JSON(200, activeSessionMetrics)
}

func (s *SessionServiceInMemory) GetCompletedSessionMetrics(c *gin.Context) {
	var completedSessionMetrics [] sessionMetrics
	result := s.metricsGorm.Find(&completedSessionMetrics)
	if result.Error != nil {
		c.String(400, result.Error.Error())
	}

	c.JSON(200, completedSessionMetrics)
}

func (s *SessionServiceInMemory) writeCompletedSessionMetrics(sessionMetrics *sessionMetrics) {
	s.metricsGorm.Create(sessionMetrics)
}