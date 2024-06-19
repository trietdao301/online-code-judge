package models

import "example/server/db"

type CreateSubmissionRequest struct {
	ProblemUUID       string
	Content           string `validate:"min=1,max=64000"`
	Language          string `validate:"max=32"`
	AuthorAccountUUID string
}
type CreateSubmissionResponse struct {
	Submission db.Submission
}

type GetSubmissionRequest struct {
	UUID string
}

type DeleteSubmissionRequest struct {
}

type UpdateSubmissionRequest struct {
	UUID          string
	GradingResult string
	Status        uint8 `validate:"oneof=1 2 3"`
	Result        uint8 `validate:"oneof=1 2 3 4 5 6 7"`
}

type GetSubmissionResponse struct {
	Submission db.Submission
}

type CreateTestCaseRequest struct {
	ProblemUUID string
	Content     string `validate:"max=5242880"`
	IsHidden    bool
}

type GetTestCaseRequest struct {
	UUID string
}
type GetTestCaseResponse struct {
	TestCase db.TestCase
}

type GetProblemRequest struct {
	UUID string
}

type GetProblemResponse struct {
	Problem db.Problem
}

type CreateProblemRequest struct {
	DisplayName            string
	Description            string
	AuthorAccountUUID      string
	TimeLimitInMillisecond uint64
	MemoryLimitInByte      uint64
}

type CreateProblemResponse struct {
	UUID              string
	DisplayName       string
	Description       string
	AuthorAccountUUID string
}
