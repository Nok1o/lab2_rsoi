package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReservationStatus string

const (
	StatusRented   ReservationStatus = "RENTED"
	StatusReturned ReservationStatus = "RETURNED"
	StatusExpired  ReservationStatus = "EXPIRED"
)

type Reservation struct {
	Id        uuid.UUID
	Username  string
	BookId    uuid.UUID
	LibraryId uuid.UUID
	Status    ReservationStatus
	StartDate time.Time
	TillDate  time.Time
}
