package logic

import (
	"context"
	"example/server/configs"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CompileOutput struct {
	ProgramFilePath string
	ReturnCode      int64
	StdOut          string
	StdErr          string
}

type Compile interface {
	Compile(ctx context.Context, content string) (CompileOutput, error)
}

type compile struct {
	dockerClient    *client.Client
	logger          *zap.Logger
	language        string
	compileConfig   *configs.Compile
	timeoutDuration time.Duration
	memoryInBytes   int64
}

func (c compile) pullImage() error {

	_, err := c.dockerClient.ImagePull(context.Background(), c.compileConfig.Image, image.PullOptions{})
	if err != nil {
		c.logger.Error("failed to pull compile image" + ": " + c.compileConfig.Image)
		return err
	}
	c.logger.Info("Successfully pulled compile image" + ": " + c.compileConfig.Image)
	return nil
}

func NewCompileLogic(docker *client.Client,
	logger *zap.Logger,
	language string,
	compileConfig *configs.Compile,
) (Compile, error) {
	c := &compile{
		dockerClient:  docker,
		logger:        logger,
		language:      language,
		compileConfig: compileConfig,
	}

	if compileConfig == nil {
		return c, nil
	}

	timeoutDuration, err := time.ParseDuration(compileConfig.Timeout)
	if err != nil {
		logger.Error("failed to parse timeout string")
	}
	c.timeoutDuration = timeoutDuration
	memoryInBytes, err := humanize.ParseBytes(compileConfig.Memory)
	if err != nil {
		c.logger.With(zap.Error(err)).Error("failed to get memory in bytes")
		return nil, err
	}
	c.memoryInBytes = int64(memoryInBytes)

	go func() {
		c.pullImage()
	}()

	return c, nil
}

// Compile implements Compile.
func (c *compile) Compile(ctx context.Context, content string) (CompileOutput, error) {
	hostWorkingDir, err := os.MkdirTemp("", "")
	if err != nil {
		c.logger.With(zap.Error(err)).Error("failed to create working directory for compiling")
		return CompileOutput{}, err
	}

	var sourceFile *os.File
	if c.compileConfig == nil {
		sourceFile, err = c.createSourceFile(ctx, hostWorkingDir, uuid.NewString(), content)
		if err != nil {
			return CompileOutput{}, err
		}

		return CompileOutput{
			ProgramFilePath: sourceFile.Name(),
		}, nil
	}

	defer func() {
		if err = os.Remove(sourceFile.Name()); err != nil {
			c.logger.With(zap.Error(err)).Error("failed to remove source file")
		}
	}()
	return CompileOutput{}, err
}

func (c compile) createSourceFile(
	ctx context.Context,
	hostWorkingDir string,
	sourceFileName string,
	content string,
) (*os.File, error) {

	sourceFilePath := filepath.Join(hostWorkingDir, sourceFileName)
	sourceFile, err := os.Create(sourceFilePath)
	c.logger.Info(sourceFileName + "is source file name")
	c.logger.Info(sourceFilePath + "is source file path")

	if err != nil {
		c.logger.With(zap.Error(err)).Error("failed to create source file")
		return nil, err
	}

	defer sourceFile.Close()

	_, err = sourceFile.WriteString(content)
	if err != nil {
		c.logger.With(zap.Error(err)).Error("failed to write source file")
		return nil, err
	}

	return sourceFile, nil
}
