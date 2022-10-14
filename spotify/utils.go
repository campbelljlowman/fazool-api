package spotify

import (
	"fmt"
)

func RefreshToken(UserID int) string {
	fmt.Printf("Refreshing Token for %v\n", UserID)
	return "New token!"
}