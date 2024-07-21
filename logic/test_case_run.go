package logic

import (
	"bytes"
	"context"
	"errors"
	"example/server/configs"
	"example/server/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// func (t testCaseRun) setupWorkingDir(workingDir string) error {
// 	if strings.ToLower(t.language) == "java" {
// 		if t.testCaseRunConfig.DownloadTestUrl != nil && t.testCaseRunConfig.TestLibraryName != nil {
// 			t.logger.Info(*t.testCaseRunConfig.DownloadTestUrl)
// 			junitURL := *t.testCaseRunConfig.DownloadTestUrl
// 			junitPath := filepath.Join(workingDir, *t.testCaseRunConfig.TestLibraryName)
// 			if err := t.downloadFile(junitURL, junitPath); err != nil {
// 				t.logger.Error(err.Error())
// 				return fmt.Errorf("fail to download test file in setupWorkingDir")
// 			}
// 		}
// 	} else if strings.ToLower(t.language) == "python" {
// 		return nil
// 	}
// 	return nil
// }

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
	//err = t.setupWorkingDir(hostWorkingDir)
	if err != nil {
		return RunOutput{}, err
	}
	resp, err := t.createContainer(ctx,
		workingDir,
		codeFile,
		testFile,
		timeLimitInSecond,
		memoryLimitInByte,
		t.testCaseRunConfig.Image,
		t.testCaseRunConfig.CommandTemplate,
		t.testCaseRunConfig.CPUQuota,
	)
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

	// timeoutDuration, err := time.ParseDuration(timeLimitInSecond)
	// if err != nil {
	// 	return RunOutput{ReturnLog: "Server Error" + err.Error()}, err
	// }
	// contextTimeout, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	// defer cancel()

	statusCh, errCh := t.dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return RunOutput{StdErr: "container channel response with error"}, err
	case <-statusCh:
		out, err := t.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			t.logger.Error("Failed to get container logs", zap.Error(err))
			return RunOutput{}, fmt.Errorf("failed to get container logs: %w", err)
		}
		defer out.Close() // Make sure to close the log output

		if out == nil {
			t.logger.Error("Container log output is nil")
			return RunOutput{}, errors.New("container log output is nil")
		}

		var stdoutBuf, stderrBuf bytes.Buffer
		_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, out)
		if err != nil {
			t.logger.Error("Failed to copy container logs", zap.Error(err))
			return RunOutput{}, fmt.Errorf("failed to copy container logs: %w", err)
		}

		stdoutLog := stdoutBuf.String()
		stderrLog := stderrBuf.String()

		if stdoutLog == "" && stderrLog == "" {
			t.logger.Warn("Container logs are empty")
			return RunOutput{}, errors.New("container logs are empty")
		}

		// t.logger.Info("Container logs retrieved",
		// 	zap.String("stdout", stdoutLog),
		// 	zap.String("stderr", stderrLog))

		returnLog, err := t.handleReturnLog(stderrLog, stdoutLog)
		if err != nil {
			return RunOutput{ReturnLog: returnLog}, nil
		}
		return RunOutput{
			ReturnLog: returnLog,
		}, nil
		// case <-contextTimeout.Done():
		// 	return RunOutput{
		// 		ReturnLog: "running submission failed: time limit exceeded > " + timeLimitInSecond,
		// 	}, nil
	}
}

func (t testCaseRun) handleReturnLog(stderrLog string, stdoutLog string) (returnLog string, err error) {
	var result []string
	if t.testCaseRunConfig.StdErr {
		result = append(result, stderrLog)
	}
	if t.testCaseRunConfig.StdOut {
		result = append(result, stdoutLog)
	}
	if len(result) == 0 {
		return "No output stream", fmt.Errorf("No container log found")
	} else {
		var result string = strings.Join(result, "\n")
		if strings.ToLower(t.language) == "java" {
			result = utils.FilterJUnitOutput(result)
		}
		return result, nil
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
	t.logger.Info(workingDir)
	t.logger.Info("CodeFile: " + codeFile.Name())
	t.logger.Info("TestFile: " + testFile.Name())
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
	if strings.ToLower(t.language) == "python" {
		for i := range commandList {
			if commandList[i] == "$TIME_LIMIT" {
				command[i] = timeLimitInSecond
			} else if commandList[i] == "$TEST_FILE" {
				command[i] = testFileName
			} else if commandList[i] == "$MAIN_FILE" {
				t.logger.Info("code file is" + (t.testCaseRunConfig.CodeFileName))
				command[i] = t.testCaseRunConfig.CodeFileName
			} else {
				command[i] = commandList[i]
			}
		}
		t.logger.Info("log cmd", zap.Any("command", command))
		return command
	} else if strings.ToLower(t.language) == "java" {
		t.logger.Info("java language:" + strings.ToLower(t.language))
		for index, item := range commandList {
			if strings.Contains(item, "$TIME_LIMIT") {
				newExecutingCommand := strings.Replace(item, "$TIME_LIMIT", timeLimitInSecond, 1)
				command[index] = newExecutingCommand
			} else {
				command[index] = commandList[index]
			}
		}
		t.logger.Info("log cmd", zap.Any("command", command))
		return command
	}
	return commandList
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

func (t testCaseRun) downloadFile(url, filepath string) error {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			t.logger.Warn("Attempt to download file failed",
				zap.Int("attempt", i+1),
				zap.Error(err),
				zap.String("url", url))
			if i == maxRetries-1 {
				return err
			}
			time.Sleep(time.Second * 2)
			continue
		}
		defer resp.Body.Close()

		out, err := os.Create(filepath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to write file content: %w", err)
		}
		t.logger.Info("File downloaded successfully", zap.String("path", filepath))
		return nil
	}
	return fmt.Errorf("failed to download file after %d attempts", maxRetries)
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
