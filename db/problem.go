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
	GetAllProblems(ctx context.Context) (*[]Problem, error)
	GetTestCaseListByProblemUUID(ctx context.Context, problemUUID string) ([]TestCaseData, error)
	DeleteProblem(ctx context.Context, problemUUID string) error
}

type problemDataAccessor struct {
	db     *mongo.Collection
	logger *zap.Logger
}

type TestCaseData struct {
	TestCaseUUID string `json:"testCaseUUID" bson:"testCaseUUID" validate:"required"`
	Language     string `json:"language" bson:"language" validate:"required"`
}
type SubmissionSnippetData struct {
	SubmissionSnippetUUID string `json:"submissionSnippetUUID" bson:"submissionSnippetUUID" validate:"required"`
	Language              string `json:"language" bson:"language" validate:"required"`
}
type Problem struct {
	UUID                   string                  `json:"UUID" bson:"UUID" validate:"required"`
	DisplayName            string                  `json:"displayName" bson:"displayName" validate:"required"`
	Description            string                  `json:"description" bson:"description" validate:"required"`
	TestCaseList           []TestCaseData          `json:"testCaseList" bson:"testCaseList"`
	AuthorAccountUUID      string                  `json:"authorAccountUUID" bson:"authorAccountUUID" validate:"required"`
	AuthorName             string                  `json:"authorName" bson:"authorName" validate:"required"`
	TimeLimitInMillisecond uint64                  `json:"timeLimitInMillisecond" bson:"timeLimitInMillisecond"`
	MemoryLimitInByte      uint64                  `json:"memoryLimitInByte" bson:"memoryLimitInByte"`
	CreatedAt              string                  `json:"createdAt" bson:"createdAt"`
	UpdatedAt              string                  `json:"updatedAt" bson:"updatedAt"`
	SubmissionSnippetList  []SubmissionSnippetData `json:"submissionSnippetList" bson:"submissionSnippetList"`
}

func (p *problemDataAccessor) DeleteProblem(ctx context.Context, problemUUID string) error {
	filter := bson.M{
		"UUID": problemUUID,
	}
	_, err := p.db.DeleteOne(ctx, filter)
	if err != nil {
		p.logger.Error("fail to delete problem", zap.Any("problemUUID", problemUUID))
		return err
	}
	return nil
}

func (p *problemDataAccessor) GetProblemByUUID(ctx context.Context, UUID string) (*Problem, error) {
	filter := bson.M{"UUID": UUID}
	var problem Problem
	err := p.db.FindOne(ctx, filter).Decode(&problem)
	if err != nil {
		p.logger.Error("fail to find problem database", zap.Error(err))
		return &Problem{}, err
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

	result, err := p.db.UpdateOne(ctx, filter, update)
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

func (p *problemDataAccessor) GetAllProblems(ctx context.Context) (*[]Problem, error) {
	var problems []Problem

	// Create a cursor for all documents in the collection
	cursor, err := p.db.Find(ctx, bson.M{})
	if err != nil {
		p.logger.Error("failed to create cursor for all problems", zap.Error(err))
		return &[]Problem{}, err
	}
	defer cursor.Close(ctx)

	// Iterate through the cursor and decode each document
	for cursor.Next(ctx) {
		var problem Problem
		if err := cursor.Decode(&problem); err != nil {
			p.logger.Error("failed to decode problem", zap.Error(err))
			return &[]Problem{}, err
		}
		problems = append(problems, problem)
	}

	// Check if the cursor encountered any errors during iteration
	if err := cursor.Err(); err != nil {
		p.logger.Error("cursor encountered an error", zap.Error(err))
		return &[]Problem{}, err
	}

	p.logger.Info("successfully retrieved all problems", zap.Int("count", len(problems)))
	return &problems, nil
}

func (p *problemDataAccessor) GetTestCaseListByProblemUUID(ctx context.Context, problemUUID string) ([]TestCaseData, error) {
	var problem Problem
	filter := bson.M{"UUID": problemUUID}
	err := p.db.FindOne(ctx, filter).Decode(&problem)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			p.logger.Warn("Problem not found", zap.String("problemUUID", problemUUID))
			return []TestCaseData{}, fmt.Errorf("problem not found: %s", problemUUID)
		}
		p.logger.Error("Failed to find problem", zap.String("problemUUID", problemUUID), zap.Error(err))
		return []TestCaseData{}, fmt.Errorf("failed to find problem: %w", err)
	}

	if len(problem.TestCaseList) == 0 {
		p.logger.Warn("No test cases found for problem", zap.String("problemUUID", problemUUID))
		return []TestCaseData{}, nil
	}

	p.logger.Info("Retrieved test cases", zap.String("problemUUID", problemUUID), zap.Int("count", len(problem.TestCaseList)))
	p.logger.Info("problem: ", zap.Any("test case", problem.TestCaseList))
	return problem.TestCaseList, nil
}

func NewProblemDataAccessor(db *mongo.Collection, logger *zap.Logger) (ProblemDataAccessor, error) {
	return &problemDataAccessor{db: db, logger: logger}, nil
}
