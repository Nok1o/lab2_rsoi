package library

import (
	"github.com/google/uuid"

	"libriary_system/shared/domain"
	"libriary_system/shared/pagination"
)

type libraryResponse struct {
	UID     uuid.UUID `json:"libraryUid"`
	Name    string    `json:"name"`
	City    string    `json:"city"`
	Address string    `json:"address"`
}

type libraryBookResponse struct {
	UID            uuid.UUID            `json:"bookUid"`
	Name           string               `json:"name"`
	Author         string               `json:"author"`
	Genre          string               `json:"genre"`
	Condition      domain.BookCondition `json:"condition"`
	AvailableCount int                  `json:"availableCount"`
}

type pageResponse[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}

type returnBookRequest struct {
	Condition domain.BookCondition `json:"condition"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func (response libraryResponse) toDomain() domain.Library {
	return domain.Library{
		Id:      response.UID,
		Name:    response.Name,
		City:    response.City,
		Address: response.Address,
	}
}

func libraryPageToDomain(
	response pageResponse[libraryResponse],
) pagination.Page[domain.Library] {
	items := make([]domain.Library, 0, len(response.Items))
	for _, library := range response.Items {
		items = append(items, library.toDomain())
	}

	return pagination.Page[domain.Library]{Items: items, Total: response.Total}
}

func (response libraryBookResponse) toDomain() domain.LibraryBook {
	return domain.LibraryBook{
		Book: domain.Book{
			Id:        response.UID,
			Name:      response.Name,
			Author:    response.Author,
			Genre:     response.Genre,
			Condition: response.Condition,
		},
		AvailableCount: response.AvailableCount,
	}
}

func libraryBookPageToDomain(
	response pageResponse[libraryBookResponse],
) pagination.Page[domain.LibraryBook] {
	items := make([]domain.LibraryBook, 0, len(response.Items))
	for _, book := range response.Items {
		items = append(items, book.toDomain())
	}

	return pagination.Page[domain.LibraryBook]{Items: items, Total: response.Total}
}
