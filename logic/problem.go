package logic

import (
	"context"
	"example/server/db"
	"example/server/handlers/models"
	"sync"
	"time"

	"example/server/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Problem interface {
	GetProblemByUUID(ctx context.Context, in *models.GetProblemRequest) (*models.GetProblemResponse, error)
	CreateProblem(ctx context.Context, in *models.CreateProblemRequest) (*models.CreateProblemResponse, error)
	GetAllProblems(ctx context.Context) (*models.GetProblemListResponse, error)
	GetAllTestCasesByProblemUUID(ctx context.Context, in *models.GetTestCaseListRequest) (*models.GetTestCaseListResponse, error)
	DeleteProblem(ctx context.Context, in *models.DeleteProblemRequest) error
}
type problem struct {
	logger                        *zap.Logger
	problemDataAccessor           db.ProblemDataAccessor
	testDataAccessor              db.TestCaseDataAccessor
	submissionSnippetDataAccessor db.SubmissionSnippetDataAccessor
}

func (p problem) deleteAllTestsInProblem(ctx context.Context, testCaseList []db.TestCaseData) error {
	for _, test := range testCaseList {
		err := p.testDataAccessor.DeleteTestCase(ctx, test.TestCaseUUID)
		if err != nil {
			p.logger.Error("fail to delete test", zap.Any("testUUID", test.TestCaseUUID))
			return err
		}
	}
	return nil
}

func (p problem) deleteAllSubmissionSnippetInProblem(ctx context.Context, submissionSnippetList []db.SubmissionSnippetData) error {
	for _, snippet := range submissionSnippetList {
		err := p.submissionSnippetDataAccessor.DeleteSubmissionSnippet(ctx, snippet.SubmissionSnippetUUID)
		if err != nil {
			p.logger.Error("fail to delete snippet", zap.Any("submissionSnippetUUID", snippet.SubmissionSnippetUUID))
			return err
		}
	}
	return nil
}

func (p problem) DeleteProblem(ctx context.Context, in *models.DeleteProblemRequest) error {
	problem, err := p.problemDataAccessor.GetProblemByUUID(ctx, in.ProblemUUID)
	if err != nil {
		p.logger.Error("fail to get problem by UUID", zap.Any("problemUUID", in.ProblemUUID))
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	var submissionSnippetListErr, testCaseListErr error

	go func() {
		defer wg.Done()
		testCaseListErr = p.deleteAllTestsInProblem(ctx, problem.TestCaseList)
	}()

	go func() {
		defer wg.Done()
		submissionSnippetListErr = p.deleteAllSubmissionSnippetInProblem(ctx, problem.SubmissionSnippetList)
	}()
	wg.Wait()
	if testCaseListErr != nil {
		return err
	}

	if submissionSnippetListErr != nil {
		return err
	}

	p.logger.Info("deleteting")
	err = p.problemDataAccessor.DeleteProblem(ctx, in.ProblemUUID)
	if err != nil {
		return err
	}
	return nil
}

func (p problem) GetProblemByUUID(ctx context.Context, in *models.GetProblemRequest) (*models.GetProblemResponse, error) {
	problem, err := p.problemDataAccessor.GetProblemByUUID(ctx, in.UUID)
	if err != nil {
		p.logger.Error("fail to get problem by uuid", zap.Error(err))
		return nil, nil
	}

	return &models.GetProblemResponse{Problem: *problem}, nil
}

func (p problem) CreateProblem(ctx context.Context, in *models.CreateProblemRequest) (*models.CreateProblemResponse, error) {
	currentTime := utils.FormatTime(time.Now())

	problem := db.Problem{
		UUID:                   uuid.NewString(),
		DisplayName:            in.DisplayName,
		Description:            in.Description,
		AuthorAccountUUID:      in.AuthorAccountUUID,
		AuthorName:             in.AuthorName,
		TimeLimitInMillisecond: in.TimeLimitInMillisecond,
		MemoryLimitInByte:      in.MemoryLimitInByte,
		CreatedAt:              currentTime,
		UpdatedAt:              currentTime,
		TestCaseList:           []db.TestCaseData{},
		SubmissionSnippetList:  []db.SubmissionSnippetData{},
	}

	err := p.problemDataAccessor.CreateProblem(ctx, &problem)
	if err != nil {
		p.logger.Error("fail to add a problem into database", zap.Any("problem", problem))
		return nil, err
	}
	p.logger.Info("Successfully created a problem", zap.Any("problem", problem))
	return &models.CreateProblemResponse{
		UUID:              problem.UUID,
		DisplayName:       problem.DisplayName,
		Description:       problem.Description,
		AuthorAccountUUID: problem.AuthorAccountUUID,
		CreatedAt:         problem.CreatedAt,
		UpdatedAt:         problem.UpdatedAt,
	}, nil
}

func (p problem) GetAllTestCasesByProblemUUID(ctx context.Context, in *models.GetTestCaseListRequest) (*models.GetTestCaseListResponse, error) {
	var list []db.TestCase
	testCaseList, err := p.problemDataAccessor.GetTestCaseListByProblemUUID(ctx, in.ProblemUUID)
	if err != nil {
		p.logger.Error("fail to get test case list by problem UUID", zap.Any("problemUUID", in.ProblemUUID))
		return &models.GetTestCaseListResponse{}, err
	}
	if len(testCaseList) == 0 {
		p.logger.Warn("test case list is 0", zap.Any("problemUUID", in.ProblemUUID))
		return &models.GetTestCaseListResponse{}, err
	}
	for _, test := range testCaseList {
		testCase, err := p.testDataAccessor.GetTestCaseByUUID(ctx, test.TestCaseUUID)
		p.logger.Info("test case: ", zap.Any("testCase", test.TestCaseUUID))
		if err != nil {
			p.logger.Error("fail to get test case by UUID", zap.Any("testUUID", test.TestCaseUUID))
			return &models.GetTestCaseListResponse{}, err
		}
		list = append(list, *testCase)
	}
	return &models.GetTestCaseListResponse{TestCaseList: list}, nil
}

func (p problem) GetAllProblems(ctx context.Context) (*models.GetProblemListResponse, error) {

	listOfProblems, err := p.problemDataAccessor.GetAllProblems(ctx)
	if err != nil {
		return nil, err
	}
	totalCount := len(*listOfProblems)
	response := &models.GetProblemListResponse{
		ListOfProblem: *listOfProblems,
		TotalCount:    totalCount}
	p.logger.Info("successfully retrieved all problems", zap.Int("count", totalCount))
	return response, nil
}

func NewProblemLogic(logger *zap.Logger,
	problemDataAccessor db.ProblemDataAccessor,
	testDataAccessor db.TestCaseDataAccessor,
	submissionSnippetDataAccessor db.SubmissionSnippetDataAccessor,
) Problem {

	return &problem{logger: logger,
		problemDataAccessor:           problemDataAccessor,
		testDataAccessor:              testDataAccessor,
		submissionSnippetDataAccessor: submissionSnippetDataAccessor,
	}
}
