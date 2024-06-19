package logic

import (
	"context"
	"example/server/db"
	"example/server/handlers/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Problem interface {
	GetProblemByUUID(ctx context.Context, in *models.GetProblemRequest) (*models.GetProblemResponse, error)
	CreateProblem(ctx context.Context, in *models.CreateProblemRequest) (*models.CreateProblemResponse, error)
}
type problem struct {
	logger              *zap.Logger
	problemDataAccessor db.ProblemDataAccessor
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

	problem := db.Problem{
		UUID:                   uuid.NewString(),
		DisplayName:            in.DisplayName,
		Description:            in.Description,
		AuthorAccountUUID:      in.AuthorAccountUUID,
		TimeLimitInMillisecond: in.TimeLimitInMillisecond,
		MemoryLimitInByte:      in.MemoryLimitInByte,
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
	}, nil
}

func NewProblemLogic(logger *zap.Logger, problemDataAccessor db.ProblemDataAccessor) Problem {
	return &problem{logger: logger, problemDataAccessor: problemDataAccessor}
}
