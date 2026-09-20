package usecase

import (
	"context"

	"github.com/google/uuid"

	"libriary_system/shared/domain"
)

type ReservationRepository interface {
	ListByUsername(ctx context.Context, username string) ([]domain.Reservation, error)
	CountByUsernameAndStatus(ctx context.Context, username string, status domain.ReservationStatus) (int, error)
	GetByUIDAndUsername(ctx context.Context, reservationUID uuid.UUID, username string) (domain.Reservation, error)
	Create(ctx context.Context, reservation domain.Reservation) error
	TransitionStatus(ctx context.Context, reservationUID uuid.UUID, username string, from, to domain.ReservationStatus) error
}
