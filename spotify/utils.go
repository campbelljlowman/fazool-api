package spotify

import (
	"fmt"

	"github.com/campbelljlowman/fazool-api/database"
	"github.com/jackc/pgx/v4/pgxpool"
)

func RefreshToken(db *pgxpool.Pool, UserID int) string {
	fmt.Printf("Refreshing Token for %v\n", UserID)
	refreshToken, err := database.GetSpotifyRefreshToken(db, UserID)
	if err != nil {
		println("Error getting spotify refresh token!")
		return ""
	}

	// Hit spotify endpoint to refresh token

	// Add tokens back to database
	return refreshToken
}