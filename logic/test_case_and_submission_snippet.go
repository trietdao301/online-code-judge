package logic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"example/server/handlers/models"
	"example/server/utils"

	"example/server/db"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type TestCaseAndSubmissionSnippet interface {
	CreateTestCaseAndSubmissionSnippet(ctx context.Context, in *models.CreateTestCaseAndSubmissionSnippetRequest) error
}

type testCaseAndSubmissionSnippet struct {
	logger                        *zap.Logger
	problemDataAccessor           db.ProblemDataAccessor
	testCaseDataAccessor          db.TestCaseDataAccessor
	submissionSnippetDataAccessor db.SubmissionSnippetDataAccessor
}

func (t *testCaseAndSubmissionSnippet) createNewTestData(newTestCaseUUID string, language string) *db.TestCaseData {
	testToAdd := &db.TestCaseData{
		TestCaseUUID: newTestCaseUUID,
		Language:     language,
	}
	return testToAdd
}

func (t *testCaseAndSubmissionSnippet) createNewSubmissionSnippetData(newSubmissionSnippetUUID string, language string) *db.SubmissionSnippetData {
	submissionSnippetToAdd := &db.SubmissionSnippetData{
		SubmissionSnippetUUID: newSubmissionSnippetUUID,
		Language:              language,
	}
	return submissionSnippetToAdd
}

func (t *testCaseAndSubmissionSnippet) checkIfCanAddMoreTestAndSnippetToProblem(problemToUpdate *db.Problem, requestedLanguage string) bool {
	for _, test := range problemToUpdate.TestCaseList {
		if test.Language == requestedLanguage {
			t.logger.Info("test case for language already exists, please remove it before adding a new one",
				zap.String("language", test.Language))
			return false
		}
	}
	for _, submissionSnippet := range problemToUpdate.SubmissionSnippetList {
		if submissionSnippet.Language == requestedLanguage {
			t.logger.Info("submission snippet for language already exists, please remove it before adding a new one",
				zap.String("language", submissionSnippet.Language))
			return false
		}
	}
	return true
}

func (t *testCaseAndSubmissionSnippet) addTestCaseToDatabase(
	ctx context.Context,
	newTestCaseUUID string,
	ofProblemUUID string,
	codeTest string,
	language string,
	currentTime string) error {

	testCase := &db.TestCase{
		UUID:            newTestCaseUUID,
		OfProblemUUID:   ofProblemUUID,
		TestFileContent: codeTest,
		Language:        language,
		CreatedAt:       currentTime,
	}
	err := t.testCaseDataAccessor.CreateTestCase(ctx, testCase)
	if err != nil {
		t.logger.Error("fail to create test case", zap.Any("test case uuid: ", newTestCaseUUID))
		return err
	}
	return nil
}

func (t *testCaseAndSubmissionSnippet) addSubmissionSnippetToDatabase(
	ctx context.Context,
	newSubmissionSnippetUUID string,
	ofProblemUUID string,
	codeSnippet string,
	language string,
	currentTime string) error {

	newSubmissionSnippet := &db.SubmissionSnippet{
		UUID:          newSubmissionSnippetUUID,
		OfProblemUUID: ofProblemUUID,
		CodeSnippet:   codeSnippet,
		Language:      language,
		CreatedAt:     currentTime,
	}
	err := t.submissionSnippetDataAccessor.CreateSubmissionSnippet(ctx, newSubmissionSnippet)
	if err != nil {
		t.logger.Error("fail to create submission snippet", zap.Any("submission snippet uuid: ", newSubmissionSnippetUUID))
		return err
	}
	return nil
}

func (t *testCaseAndSubmissionSnippet) createUpdateFilter(
	newSubmissionSnippetUUID string,
	newTestCaseUUID string,
	language string,
	SubmissionSnippetListFieldName string,
	TestCaseListFieldName string) primitive.M {

	SubmissionSnippetListFieldName = utils.GetFieldName(db.Problem{}, SubmissionSnippetListFieldName)
	TestCaseListFieldName = utils.GetFieldName(db.Problem{}, TestCaseListFieldName)

	submissionSnippetToAdd := t.createNewSubmissionSnippetData(newSubmissionSnippetUUID, language)
	testToAdd := t.createNewTestData(newTestCaseUUID, language)

	update := bson.M{
		"$push": bson.M{
			TestCaseListFieldName:          testToAdd,
			SubmissionSnippetListFieldName: submissionSnippetToAdd,
		},
	}
	return update
}

func (t *testCaseAndSubmissionSnippet) CreateTestCaseAndSubmissionSnippet(ctx context.Context, in *models.CreateTestCaseAndSubmissionSnippetRequest) error {
	newTestCaseUUID := uuid.NewString()
	newSubmissionSnippetUUID := uuid.NewString()

	problemToUpdate, err := t.problemDataAccessor.GetProblemByUUID(ctx, in.OfProblemUUID)
	if err != nil {
		t.logger.Warn("fail to get problem by uuid", zap.Error(err))
		return fmt.Errorf("failed to get problem: %w", err)
	}

	if !t.checkIfCanAddMoreTestAndSnippetToProblem(problemToUpdate, in.Language) {
		return fmt.Errorf("problem is not available for adding more tests or snippet")
	}

	currentTime := utils.FormatTime(time.Now())

	var wg sync.WaitGroup
	wg.Add(2)

	var submissionErr, testCaseErr error

	go func() {
		defer wg.Done()
		submissionErr = t.addSubmissionSnippetToDatabase(ctx, newSubmissionSnippetUUID, in.OfProblemUUID, in.CodeSnippet, in.Language, currentTime)
	}()

	go func() {
		defer wg.Done()
		testCaseErr = t.addTestCaseToDatabase(ctx, newTestCaseUUID, in.OfProblemUUID, in.CodeTest, in.Language, currentTime)
	}()

	wg.Wait()

	if submissionErr != nil {
		return fmt.Errorf("failed to add submission snippet: %w", submissionErr)
	}

	if testCaseErr != nil {
		return fmt.Errorf("failed to add test case: %w", testCaseErr)
	}

	update := t.createUpdateFilter(newSubmissionSnippetUUID, newTestCaseUUID, in.Language, "SubmissionSnippetList", "TestCaseList")
	if err := t.problemDataAccessor.UpdateProblem(ctx, in.OfProblemUUID, update); err != nil {
		t.logger.Error("failed to update problem with new test case and submission snippet",
			zap.String("problemUUID", in.OfProblemUUID),
			zap.Error(err))
		return fmt.Errorf("failed to update problem: %w", err)
	}

	return nil
}

func NewTestCaseAndSubmissionSnippetLogic(
	logger *zap.Logger,
	problemDatAccessor db.ProblemDataAccessor,
	testCaseDataAccessor db.TestCaseDataAccessor,
	submissionSnippetDataAccessor db.SubmissionSnippetDataAccessor,
) TestCaseAndSubmissionSnippet {
	return &testCaseAndSubmissionSnippet{
		logger:                        logger,
		problemDataAccessor:           problemDatAccessor,
		testCaseDataAccessor:          testCaseDataAccessor,
		submissionSnippetDataAccessor: submissionSnippetDataAccessor,
	}
}
