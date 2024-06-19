package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type ProblemDataAccessor interface {
	CreateProblem(ctx context.Context, problem *Problem) error
	GetProblemByUUID(ctx context.Context, UUID string) (*Problem, error)
	UpdateProblem(ctx context.Context, problemUUID string, update bson.M) error
}

type problemDataAccessor struct {
	db     *mongo.Collection
	logger *zap.Logger
}

type Problem struct {
	UUID                   string `json:"UUID" bson:"UUID" validate:"required"`
	DisplayName            string `json:"displayName" bson:"displayName" validate:"required"`
	Description            string `json:"description" bson:"description" validate:"required"`
	TestCaseUUID           string `json:"testCaseUUID" bson:"testCaseUUID"`
	AuthorAccountUUID      string `json:"authorAccountUUID" bson:"authorAccountUUID" validate:"required"`
	TimeLimitInMillisecond uint64 `json:"timeLimitInMillisecond" bson:"timeLimitInMillisecond"`
	MemoryLimitInByte      uint64 `json:"memoryLimitInByte" bson:"memoryLimitInByte"`
}

func (p *problemDataAccessor) GetProblemByUUID(ctx context.Context, UUID string) (*Problem, error) {
	filter := bson.M{"UUID": UUID}
	var problem Problem
	err := p.db.FindOne(ctx, filter).Decode(&problem)
	if err != nil {
		p.logger.Error("fail to find problem database", zap.Error(err))
		return nil, err
	}
	return &problem, nil
}

func (p *problemDataAccessor) CreateProblem(ctx context.Context, problem *Problem) error {
	_, err := p.db.InsertOne(ctx, problem)
	if err != nil {
		p.logger.Error("fail to create problem in database")
		return err
	}
	return nil
}

func (p *problemDataAccessor) UpdateProblem(ctx context.Context, problemUUID string, update bson.M) error {
	filter := bson.M{"UUID": problemUUID}

	// Create an update document using $set operator
	updateDoc := bson.M{
		"$set": update,
	}
	result, err := p.db.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		p.logger.Error("failed to update problem", zap.String("problemUUID", problemUUID), zap.Error(err))
		return err
	}
	if result.MatchedCount == 0 {
		p.logger.Warn("no problem found to update", zap.String("problemUUID", problemUUID))
		return fmt.Errorf("no problem found with UUID: %s", problemUUID)
	}
	p.logger.Info("problem updated successfully", zap.String("problemUUID", problemUUID))
	return nil
}

func NewProblemDataAccessor(db *mongo.Collection, logger *zap.Logger) (ProblemDataAccessor, error) {
	return &problemDataAccessor{db: db, logger: logger}, nil
}
