package validation

import (
	"fmt"
	"libriary_system/shared/pagination"
)

const maxPageSize = 100

func ValidatePageToken(pageToken pagination.PageToken) map[string]string {
	fields := make(map[string]string)
	if pageToken.Limit < 1 || pageToken.Limit > maxPageSize {
		fields["limit"] = fmt.Sprintf("must be between 1 and %d", maxPageSize)
	}
	if pageToken.Offset < 0 {
		fields["offset"] = "must be greater than or equal to zero"
	}

	return fields
}
