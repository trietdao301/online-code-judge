package models

import (
	"example/server/db"
)

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

type GetSubmissionListResponse struct {
	Submissions []*db.Submission
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
	Language    string
}
type GetTestCaseListRequest struct {
	ProblemUUID string
}
type GetTestCaseListResponse struct {
	TestCaseList []db.TestCase
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
	AuthorName             string
	TimeLimitInMillisecond uint64
	MemoryLimitInByte      uint64
}

type DeleteProblemRequest struct {
	ProblemUUID string
}

type CreateProblemResponse struct {
	UUID              string
	DisplayName       string
	Description       string
	AuthorAccountUUID string
	CreatedAt         string
	UpdatedAt         string
}

type GetProblemListResponse struct {
	ListOfProblem []db.Problem
	TotalCount    int
}

type GetSubmissionSnippetRequest struct {
	SubmissionSnippetUUID string
}

type GetSubmissionSnippetResponse struct {
	CodeSnippet string
	Language    string
}

type CreateSubmissionSnippetRequest struct {
	CodeSnippet   string
	Language      string
	OfProblemUUID string
}

type CreateSubmissionSnippetResponse struct {
	UUID     string
	Language string
}

type CreateTestCaseAndSubmissionSnippetRequest struct {
	CodeSnippet   string
	CodeTest      string
	OfProblemUUID string
	Language      string
}

type GetAccountRequest struct {
	UUID string
}

type GetAccountResponse struct {
	Account db.Account
}
type CreateAccountRequest struct {
	Username string
	Password string
	Role     string
}
type CreateAccountResponse struct {
	Username string
	Role     string
}

type GetAccountListResponse struct {
	ListOfAccounts []db.Account
	TotalCount     int
}

type UpdateAccountRequest struct {
	UUID           string
	RequestingRole string
}

type UpdateAccountResponse struct {
	UUID      string
	Username  string
	Role      string
	UpdatedAt string
}

type DeleteAccountRequest struct {
	UUID string
}

type CreateSessionRequest struct {
	Username string
	Password string
}

type CreateSessionResponse struct {
	Token       string
	Username    string
	Role        string
	AccountUUID string
}
