package usecase

import (
	"libriary_system/shared/domain"
	"libriary_system/shared/validation"
)

func validateRating(rating domain.Rating) error {
	if rating.StarsCount < minStars || rating.StarsCount > maxStars {
		return InvalidStarsCountErr
	}

	return validation.ValidateUsername(rating.Username)
}
