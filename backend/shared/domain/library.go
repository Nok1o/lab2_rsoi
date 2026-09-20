package domain

import "github.com/google/uuid"

type BookCondition string

const (
	ConditionExcellent BookCondition = "EXCELLENT"
	ConditionGood      BookCondition = "GOOD"
	ConditionBad       BookCondition = "BAD"
)

type Library struct {
	Id      uuid.UUID
	Name    string
	City    string
	Address string
}

type Book struct {
	Id        uuid.UUID
	Name      string
	Author    string
	Genre     string
	Condition BookCondition
}

type LibraryBook struct {
	Book
	AvailableCount int
}
