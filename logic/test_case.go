package logic

import (
	"context"
	"fmt"
	"time"

	"example/server/db"
	"example/server/handlers/models"
	"example/server/utils"

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
	if testData == nil {
		return &models.GetTestCaseResponse{}, err
	}
	if err != nil {
		return &models.GetTestCaseResponse{}, err
	}
	return &models.GetTestCaseResponse{TestCase: *testData}, nil
}

func (t testCase) CreateTestCase(ctx context.Context, in *models.CreateTestCaseRequest) error {
	var (
		testToAdd       *db.TestCaseData
		UUID            string
		problemToUpdate *db.Problem
	)
	UUID = uuid.NewString()
	testToAdd = &db.TestCaseData{
		TestCaseUUID: UUID,
		Language:     in.Language,
	}

	TestCasesFieldName := utils.GetFieldName(db.Problem{}, "TestCaseList")
	update := bson.M{
		"$push": bson.M{
			TestCasesFieldName: testToAdd,
		},
	}

	problemToUpdate, err := t.problemDataAccessor.GetProblemByUUID(ctx, in.ProblemUUID)
	if err != nil {
		t.logger.Error("fail to get problem by uuid")
	}
	for _, test := range problemToUpdate.TestCaseList {
		if test.Language == in.Language {
			t.logger.Info("test case for language already exists, please remove it before adding a new one",
				zap.String("language", test.Language))
			return fmt.Errorf("test case for language %s already exists, please remove it before adding a new one", test.Language)
		}
	}

	currentTime := utils.FormatTime(time.Now())
	testCase := db.TestCase{
		UUID:            UUID,
		OfProblemUUID:   in.ProblemUUID,
		TestFileContent: in.Content,
		Language:        in.Language,
		CreatedAt:       currentTime,
	}

	err = t.testCaseDataAccessor.CreateTestCase(ctx, &testCase)
	if err != nil {
		return err
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

func NewTestCaseLogic(testCaseDataAccessor db.TestCaseDataAccessor, problemDataAccessor db.ProblemDataAccessor, logger *zap.Logger) (t TestCase) {
	return &testCase{testCaseDataAccessor: testCaseDataAccessor, problemDataAccessor: problemDataAccessor, logger: logger}
}
