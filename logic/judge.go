package logic

import (
	"context"
	"example/server/configs"

	"example/server/db"

	"github.com/docker/docker/client"
	"github.com/gammazero/workerpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Judge interface {
	JudgeSubmission(ctx context.Context, language string, content string, test_file string)
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

func NewJudgeLogic(logger *zap.Logger, db *mongo.Client, docker *client.Client, judgeConfig *configs.Judge) (Judge, error) {

	j := &judge{
		logger:                     logger,
		workerPool:                 workerpool.New(1),
		db:                         db,
		judgeConfig:                judgeConfig,
		languageToTestCaseRunLogic: make(map[string]TestCaseRun),
	}

	for _, language := range judgeConfig.Languages {
		testCaseRun, err := NewTestCaseRunLogic(docker, logger, language.Value, &language.TestCaseRun)
		if err != nil {
			logger.Error("fail to make new test case logic")
		}
		j.languageToTestCaseRunLogic[language.Value] = testCaseRun
	}
	return j, nil
}

func (j judge) JudgeSubmission(ctx context.Context, language string, content string, test_file string) {

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
	testCase, err := j.testDataAccessor.GetTestCaseByProblemUUID(ctx, problem.UUID)
	if err != nil {
		j.logger.Error("fail to get test case by problemUUID", zap.Error(err), zap.Any("problemUUID", problem.UUID))
		return
	}
	j.JudgeSubmission(ctx, submissionDB.Language, submissionDB.Content, testCase.TestFileContent)

}

func (j judge) ScheduleJudgeLocalSubmission(submissionUUID string) {
	j.workerPool.Submit(func() { j.judgeLocalSubmission(context.Background(), submissionUUID) })
}
