package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type SubmissionSnippetDataAccessor interface {
	GetSubmissionSnippetByUUID(ctx context.Context, submissionSnippetUUID string) (*SubmissionSnippet, error)
	CreateSubmissionSnippet(ctx context.Context, submissionSnippet *SubmissionSnippet) error
	DeleteSubmissionSnippet(ctx context.Context, submissionSnippetUUID string) error
}
type submissionSnippetDataAccessor struct {
	db     *mongo.Collection
	logger *zap.Logger
}
type SubmissionSnippet struct {
	UUID          string `json:"UUID" bson:"UUID" validate:"required"`
	OfProblemUUID string `json:"ofProblemUUID" bson:"ofProblemUUID" validate:"required"`
	CodeSnippet   string `json:"codeSnippet" bson:"codeSnippet" validate:"required"`
	Language      string `json:"language" bson:"language" validate:"required"`
	CreatedAt     string `json:"createdAt" bson:"createdAt"`
}

func (s *submissionSnippetDataAccessor) DeleteSubmissionSnippet(ctx context.Context, submissionSnippetUUID string) error {
	filter := bson.M{
		"UUID": submissionSnippetUUID,
	}
	_, err := s.db.DeleteOne(ctx, filter)
	if err != nil {
		s.logger.Error("fail to delete submission snippet", zap.Any("submissionSnippetUUID", submissionSnippetUUID), zap.Any("error: ", err.Error()))
		return err
	}
	return nil
}

func (s *submissionSnippetDataAccessor) GetSubmissionSnippetByUUID(ctx context.Context, submissionSnippetUUID string) (*SubmissionSnippet, error) {
	var (
		snippet SubmissionSnippet
	)
	filter := bson.M{"UUID": submissionSnippetUUID}
	err := s.db.FindOne(ctx, filter).Decode(&snippet)
	if err != nil {
		s.logger.Error("fail to find code snippet", zap.Any("submission snippet uuid:", submissionSnippetUUID))
		return &SubmissionSnippet{}, err
	}
	return &snippet, nil
}

func (s *submissionSnippetDataAccessor) CreateSubmissionSnippet(ctx context.Context, submissionSnippet *SubmissionSnippet) error {
	_, err := s.db.InsertOne(ctx, submissionSnippet)
	if err != nil {
		s.logger.Error("failed to create submission snippet", zap.Error(err), zap.Any("snippet", submissionSnippet))
		return err
	}
	s.logger.Info("submission snippet created successfully", zap.String("UUID", submissionSnippet.UUID))
	return nil
}

func NewSubmissionSnippetDataAccessor(db *mongo.Collection, logger *zap.Logger) SubmissionSnippetDataAccessor {
	return &submissionSnippetDataAccessor{db: db, logger: logger}
}
