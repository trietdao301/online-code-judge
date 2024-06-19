package logic

import (
	"context"

	"example/server/db"
	"example/server/handlers/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type TestCase interface {
	CreateTestCase(ctx context.Context, in *models.CreateTestCaseRequest) error
	GetTestCaseByUUID(ctx context.Context, in *models.GetTestCaseRequest) (*models.GetTestCaseResponse, error)
	GetTestCaseByProblemUUID(ctx context.Context, UUID string) (*models.GetTestCaseResponse, error)
}

type testCase struct {
	testCaseDataAccessor db.TestCaseDataAccessor
	logger               *zap.Logger
	problemDataAccessor  db.ProblemDataAccessor
}

// GetTestCaseByProblemUUID implements TestCase.
func (t *testCase) GetTestCaseByProblemUUID(ctx context.Context, UUID string) (*models.GetTestCaseResponse, error) {
	testData, err := t.testCaseDataAccessor.GetTestCaseByProblemUUID(ctx, UUID)
	if err != nil {
		return &models.GetTestCaseResponse{}, err
	}
	return &models.GetTestCaseResponse{TestCase: *testData}, nil
}

// GetTestCaseByUUID implements TestCase.
func (t *testCase) GetTestCaseByUUID(ctx context.Context, in *models.GetTestCaseRequest) (*models.GetTestCaseResponse, error) {
	testData, err := t.testCaseDataAccessor.GetTestCaseByUUID(ctx, in.UUID)
	if err != nil {
		return &models.GetTestCaseResponse{}, err
	}
	return &models.GetTestCaseResponse{TestCase: *testData}, nil
}

func NewTestCaseLogic(testCaseDataAccessor db.TestCaseDataAccessor, problemDataAccessor db.ProblemDataAccessor, logger *zap.Logger) (t TestCase) {
	return &testCase{testCaseDataAccessor: testCaseDataAccessor, problemDataAccessor: problemDataAccessor, logger: logger}
}

func (t testCase) CreateTestCase(ctx context.Context, in *models.CreateTestCaseRequest) error {
	testCase := db.TestCase{UUID: uuid.NewString(), OfProblemUUID: in.ProblemUUID, TestFileContent: in.Content}

	err := t.testCaseDataAccessor.CreateTestCase(ctx, &testCase)
	if err != nil {
		return err
	}

	update := bson.M{
		"TestCaseUUID": testCase.UUID,
	}

	err = t.problemDataAccessor.UpdateProblem(ctx, in.ProblemUUID, update)
	if err != nil {
		t.logger.Error("failed to update problem with new test case",
			zap.String("problemUUID", in.ProblemUUID),
			zap.String("testCaseUUID", testCase.UUID),
			zap.Error(err))
		return err
	}

	return nil
}
