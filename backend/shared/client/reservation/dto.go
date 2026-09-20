package reservation

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"libriary_system/shared/domain"
)

const dateLayout = "2006-01-02"

type rentRequest struct {
	BookUID    uuid.UUID `json:"bookUid"`
	LibraryUID uuid.UUID `json:"libraryUid"`
	TillDate   string    `json:"tillDate"`
}

type returnRequest struct {
	Date string `json:"date"`
}

type reservationResponse struct {
	UID        uuid.UUID                `json:"reservationUid"`
	BookUID    uuid.UUID                `json:"bookUid"`
	LibraryUID uuid.UUID                `json:"libraryUid"`
	Status     domain.ReservationStatus `json:"status"`
	StartDate  string                   `json:"startDate"`
	TillDate   string                   `json:"tillDate"`
}

type countResponse struct {
	Count int `json:"count"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func (response reservationResponse) toDomain(username string) (domain.Reservation, error) {
	startDate, err := time.Parse(dateLayout, response.StartDate)
	if err != nil {
		return domain.Reservation{}, fmt.Errorf("parse reservation start date: %w", err)
	}
	tillDate, err := time.Parse(dateLayout, response.TillDate)
	if err != nil {
		return domain.Reservation{}, fmt.Errorf("parse reservation till date: %w", err)
	}
	return domain.Reservation{
		Id:        response.UID,
		Username:  username,
		BookId:    response.BookUID,
		LibraryId: response.LibraryUID,
		Status:    response.Status,
		StartDate: startDate,
		TillDate:  tillDate,
	}, nil
}
