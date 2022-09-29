// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
)

type NewUser struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type Session struct {
	ID               int     `json:"id"`
	CurrentlyPlaying *Song   `json:"currentlyPlaying"`
	Queue            []*Song `json:"queue"`
}

type Song struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Image  string `json:"image"`
	Votes  int    `json:"votes"`
}

type SongUpdate struct {
	ID     string  `json:"id"`
	Title  *string `json:"title"`
	Artist *string `json:"artist"`
	Image  *string `json:"image"`
	Vote   int     `json:"vote"`
}

type Token struct {
	Jwt string `json:"jwt"`
}

type User struct {
	ID        int     `json:"id"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email"`
	SessionID *int    `json:"sessionID"`
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type QueueAction string

const (
	QueueActionPlay    QueueAction = "PLAY"
	QueueActionPause   QueueAction = "PAUSE"
	QueueActionAdvance QueueAction = "ADVANCE"
)

var AllQueueAction = []QueueAction{
	QueueActionPlay,
	QueueActionPause,
	QueueActionAdvance,
}

func (e QueueAction) IsValid() bool {
	switch e {
	case QueueActionPlay, QueueActionPause, QueueActionAdvance:
		return true
	}
	return false
}

func (e QueueAction) String() string {
	return string(e)
}

func (e *QueueAction) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = QueueAction(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid QueueAction", str)
	}
	return nil
}

func (e QueueAction) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
