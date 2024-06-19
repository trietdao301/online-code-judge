package logic

import (
	"context"
	"example/server/configs"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type RunOutput struct {
	ReturnLog           string
	TimeLimitExceeded   bool
	MemoryLimitExceeded bool
	StdOut              string
	StdErr              string
}

type TestCaseRun interface {
	Run(ctx context.Context,
		testFile string,
		codeFile string,
		timeLimitInMillisecond uint64,
		memoryLimitInByte uint64) (RunOutput, error)
}

type testCaseRun struct {
	dockerClient      *client.Client
	logger            *zap.Logger
	language          string
	testCaseRunConfig *configs.TestCaseRun
}

// Run implements TestCaseRun.
func (t testCaseRun) Run(ctx context.Context, testFile string, codeFile string, timeLimitInMillisecond uint64, memoryLimitInByte uint64) (RunOutput, error) {
	panic("unimplemented")
}

func (t testCaseRun) pullImage() error {
	t.logger.Info("pulling test case run image")
	_, err := t.dockerClient.ImagePull(context.Background(), t.testCaseRunConfig.Image, image.PullOptions{})
	if err != nil {
		t.logger.With(zap.Error(err)).Error("failed to pull test case run image")
		return err
	}

	t.logger.Info("pulled test case run image successfully")
	return nil
}

func NewTestCaseRunLogic(docker *client.Client, logger *zap.Logger, language string, testCaseRunConfig *configs.TestCaseRun) (TestCaseRun, error) {

	t := &testCaseRun{dockerClient: docker, logger: logger, language: language, testCaseRunConfig: testCaseRunConfig}
	if testCaseRunConfig == nil {
		t.logger.Error("fail to find test case run config")
		return t, nil
	}
	go func() {
		t.pullImage()
	}()
	return t, nil
}
