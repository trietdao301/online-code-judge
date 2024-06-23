package logic

import (
	"context"
	"errors"
	"example/server/configs"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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
		testCodeSnippet string,
		submissionCodeSnippet string,
		timeLimitInSecond string,
		memoryLimitInByte uint64) (RunOutput, error)
}

type testCaseRun struct {
	dockerClient      *client.Client
	logger            *zap.Logger
	language          string
	testCaseRunConfig *configs.TestCaseRun
}

func (t testCaseRun) Run(ctx context.Context, testCodeSnippet string, submissionCodeSnippet string, timeLimitInSecond string, memoryLimitInByte uint64) (RunOutput, error) {
	hostWorkingDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.logger.Error("fail to make temp directory", zap.Error(err))
		return RunOutput{}, err
	}
	codeFile, err := t.createTempCodeFile(ctx, hostWorkingDir, t.testCaseRunConfig.CodeFileName, submissionCodeSnippet)
	if err != nil {
		t.logger.Error("fail to create temporary code file", zap.Error(err))
		return RunOutput{}, err
	}
	testFile, err := t.createTempCodeFile(ctx, hostWorkingDir, t.testCaseRunConfig.TestFileName, testCodeSnippet)
	if err != nil {
		t.logger.Error("fail to create temporary test file", zap.Error(err))
		return RunOutput{}, err
	}
	workingDir := t.getWorkingDir()
	resp, err := t.createContainer(ctx,
		workingDir,
		codeFile,
		testFile,
		timeLimitInSecond,
		memoryLimitInByte,
		t.testCaseRunConfig.Image,
		t.testCaseRunConfig.CommandTemplate,
		t.testCaseRunConfig.CPUQuota)
	if err != nil {
		t.logger.Error("fail to create Container", zap.Any("err", err))
		return RunOutput{}, err
	}

	if err := t.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		t.logger.Error("fail to start the container", zap.Any("containerID", resp.ID))
		return RunOutput{}, err
	}

	defer func() {
		err = t.dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		if err != nil {
			t.logger.With(zap.Error(err)).Error("failed to remove run test case container")
		}
	}()

	statusCh, errCh := t.dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return RunOutput{StdErr: "container channel response with error"}, err
	case <-statusCh:
		out, err := t.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			panic(err)
		}
		buf := new(strings.Builder)
		_, err = stdcopy.StdCopy(nil, buf, out)
		if err != nil {
			t.logger.Error("fail to stdcopy the result log of the container")
			return RunOutput{}, err
		}
		containerLog := buf.String()
		if containerLog == "" {
			t.logger.Warn("Container log is empty")
			return RunOutput{}, errors.New("container log is empty")
		}

		t.logger.Info("Container log retrieved", zap.String("log", containerLog))
		return RunOutput{ReturnLog: containerLog}, nil
	}
}

func (t testCaseRun) createContainer(
	ctx context.Context,
	workingDir string,
	codeFile *os.File,
	testFile *os.File,
	timeOutOfContainerInSecond string,
	memoryLimitInByte uint64,
	image string,
	commandTemplate []string,
	CPUquota int64,
) (container.CreateResponse, error) {
	programFileDirectory, codePath := filepath.Split(codeFile.Name())
	_, testPath := filepath.Split(testFile.Name())
	filepath.Join(workingDir, codePath)
	filepath.Join(workingDir, testPath)
	resp, err := t.dockerClient.ContainerCreate(ctx, &container.Config{
		Image:      image,
		Cmd:        t.getContainerCommand(commandTemplate, timeOutOfContainerInSecond, t.testCaseRunConfig.TestFileName),
		WorkingDir: workingDir,
	}, &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:%s", programFileDirectory, workingDir)},
		Resources: container.Resources{
			CPUQuota: CPUquota,
			Memory:   int64(memoryLimitInByte)}, /// 256 * 1024 * 1024 = 256 MB
		NetworkMode: "host"}, nil, nil, "")
	if err != nil {
		return container.CreateResponse{}, err
	}
	return resp, nil
}

func (t testCaseRun) createTempCodeFile(ctx context.Context, hostWorkingDir string, sourceFileName string, content string) (*os.File, error) {
	codeFilePath := filepath.Join(hostWorkingDir, sourceFileName)
	codeFile, err := os.Create(codeFilePath)
	if err != nil {
		t.logger.Error("failed to create source file", zap.Error(err))
		return nil, err
	}
	defer codeFile.Close()
	_, err = codeFile.WriteString(content)
	if err != nil {
		t.logger.Error("failed to write source file")
		return nil, err
	}
	return codeFile, nil
}

func (t testCaseRun) getWorkingDir() string {
	return "/work"
}

func (t testCaseRun) getContainerCommand(commandList []string, timeLimitInSecond string, testFileName string) []string {
	command := make([]string, len(commandList))
	for i := range commandList {
		if commandList[i] == "$TIME_LIMIT" {
			command[i] = timeLimitInSecond
		} else if commandList[i] == "$TEST_FILE" {
			command[i] = testFileName
		} else {
			command[i] = commandList[i]
		}
	}
	t.logger.Info("log cmd", zap.Any("command", command))
	return command
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
