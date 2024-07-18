package logic

import (
	"context"
	"example/server/configs"
	"fmt"

	"example/server/db"

	"github.com/docker/docker/client"
	"github.com/gammazero/workerpool"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Judge interface {
	ScheduleJudgeLocalSubmission(submissionUUID string)
}

type judge struct {
	logger                     *zap.Logger
	workerPool                 *workerpool.WorkerPool
	db                         *mongo.Client
	languageToTestCaseRunLogic map[string]TestCaseRun
	judgeConfig                *configs.Judge
	submissionDataAccessor     db.SubmissionDataAccessor
	testDataAccessor           db.TestCaseDataAccessor
	problemDataAccessor        db.ProblemDataAccessor
}

func NewJudgeLogic(logger *zap.Logger,
	db *mongo.Client,
	docker *client.Client,
	judgeConfig *configs.Judge,
	submissionDataAccessor db.SubmissionDataAccessor,
	testDataAcessor db.TestCaseDataAccessor,
	problemDataAccessor db.ProblemDataAccessor,

) (Judge, error) {

	j := &judge{
		logger:                     logger,
		workerPool:                 workerpool.New(1),
		db:                         db,
		judgeConfig:                judgeConfig,
		languageToTestCaseRunLogic: make(map[string]TestCaseRun),
		submissionDataAccessor:     submissionDataAccessor,
		testDataAccessor:           testDataAcessor,
		problemDataAccessor:        problemDataAccessor,
	}

	for _, language := range judgeConfig.Languages {

		testCaseRun, err := NewTestCaseRunLogic(docker, logger, language.Value, &language.TestCaseRun)
		if err != nil {
			logger.Error("fail to make new test case logic")
		}
		j.languageToTestCaseRunLogic[language.Value] = testCaseRun
		logger.Info("created test run logic", zap.Any("language", language.Name))
	}

	return j, nil
}

func (j judge) judgeSubmission(
	ctx context.Context,
	language string,
	submissionCodeSnippet string,
	testCodeSnippet string,
	timeLimitInMillisecond uint64,
	memoryInByte uint64) (RunOutput, error) {

	timeLimitInSecond := fmt.Sprintf("%.3fs", float64(timeLimitInMillisecond)/1000)
	j.logger.Info("getting timeout", zap.Any("timeoutinMiniSecond", timeLimitInMillisecond), zap.Any("timeoutInSecond", timeLimitInSecond))
	if j.languageToTestCaseRunLogic[language] == nil {
		j.logger.Error("nil test case logic", zap.Any("test case language", language))
	}
	output, err := j.languageToTestCaseRunLogic[language].Run(ctx, testCodeSnippet, submissionCodeSnippet, timeLimitInSecond, memoryInByte)
	if err != nil {
		return RunOutput{}, err
	}
	if output.ReturnLog == "" {
		fmt.Printf("nil output1")
	}
	return output, nil
}

func (j judge) judgeLocalSubmission(ctx context.Context, submissionUUID string) {
	submissionDB, err := j.submissionDataAccessor.GetSubmissionByUUID(ctx, submissionUUID)
	if err != nil {
		j.logger.Error("fail to get submissionUUID", zap.Error(err))
		return
	}
	problem, err := j.problemDataAccessor.GetProblemByUUID(ctx, submissionDB.ProblemUUID)
	if err != nil {
		j.logger.Error("fail to get problem by UUID", zap.Error(err), zap.Any("problemUUID", submissionDB.ProblemUUID))
		return
	}
	testCase, err := j.testDataAccessor.GetTestCaseByProblemUUIDAndLanguage(ctx, problem.UUID, submissionDB.Language)
	if err != nil {
		j.logger.Error("fail to get test case by problemUUID", zap.Error(err), zap.Any("problemUUID", problem.UUID))
		return
	}

	output, err := j.judgeSubmission(ctx, submissionDB.Language, submissionDB.Content, testCase.TestFileContent, problem.TimeLimitInMillisecond, problem.MemoryLimitInByte)
	if err != nil {
		j.logger.Error(err.Error())
		return
	}

	j.updateSubmission(ctx, submissionUUID, output.ReturnLog, db.SubmissionStatusFinished)
}

func (j judge) updateSubmission(ctx context.Context, uuid string, gradingResult string, status db.SubmissionStatus) error {

	update := map[string]any{
		"grading_result": gradingResult,
		"status":         status,
	}
	err := j.submissionDataAccessor.UpdateSubmissionByUUID(ctx, uuid, update)
	if err != nil {
		return err
	}
	return nil
}

func (j judge) ScheduleJudgeLocalSubmission(submissionUUID string) {
	j.workerPool.Submit(func() { j.judgeLocalSubmission(context.Background(), submissionUUID) })
}
