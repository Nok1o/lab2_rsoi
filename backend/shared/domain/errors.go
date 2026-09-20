package domain

import "errors"

var (
	ErrLibraryNotFound      = errors.New("library not found")
	ErrBookNotFound         = errors.New("book not found in library")
	ErrBookUnavailable      = errors.New("book is not available")
	ErrRatingNotFound       = errors.New("rating not found")
	ErrRatingAlreadyExists  = errors.New("rating already exists")
	ErrReservationNotFound  = errors.New("reservation not found")
	ErrReservationNotRented = errors.New("reservation is not rented")
)
