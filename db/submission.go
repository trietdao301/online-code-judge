package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type SubmissionStatus uint8
type SubmissionResult uint8

const (
	SubmissionStatusSubmitted SubmissionStatus = 1
	SubmissionStatusExecuting SubmissionStatus = 2
	SubmissionStatusFinished  SubmissionStatus = 3

	SubmissionResultOK                  SubmissionResult = 1
	SubmissionResultCompileError        SubmissionResult = 2
	SubmissionResultRuntimeError        SubmissionResult = 3
	SubmissionResultTimeLimitExceeded   SubmissionResult = 4
	SubmissionResultMemoryLimitExceed   SubmissionResult = 5
	SubmissionResultWrongAnswer         SubmissionResult = 6
	SubmissionResultUnsupportedLanguage SubmissionResult = 7
)

type Submission struct {
	UUID              string           `json:"UUID" bson:"UUID"`
	ProblemUUID       string           `json:"problemUUID" bson:"problemUUID" validate:"required"`
	AuthorAccountUUID string           `json:"authorAccountUUID" bson:"authorAccountUUID" validate:"required"`
	Content           string           `json:"content" bson:"content" validate:"required,min=1,max=64000"`
	Language          string           `json:"language" bson:"language" validate:"required,max=32"`
	Status            SubmissionStatus `json:"status" bson:"status" validate:"required,oneof=1 2 3"`
	Result            SubmissionResult `json:"result" bson:"result" validate:"oneof=1 2 3 4 5 6 7"`
	GradingResult     string           `json:"grading_result" bson:"grading_result"`
	CreatedTime       int64            `json:"created_time" bson:"created_time"`
}

type submissionDataAccessor struct {
	db     *mongo.Collection
	logger *zap.Logger
}
type SubmissionDataAccessor interface {
	CreateSubmission(ctx context.Context, submission *Submission) error
	GetSubmissionByUUID(ctx context.Context, uuid string) (*Submission, error)
}

func (s *submissionDataAccessor) CreateSubmission(ctx context.Context, submission *Submission) error {
	_, err := s.db.InsertOne(ctx, submission)
	if err != nil {
		s.logger.Error("fail to create submission in database")
		return err
	}
	return nil
}

func (s *submissionDataAccessor) GetSubmissionByUUID(ctx context.Context, uuid string) (*Submission, error) {
	filter := bson.M{"UUID": uuid}
	var submission Submission
	err := s.db.FindOne(ctx, filter).Decode(&submission)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Return nil if no document is found
		}
		s.logger.Error("Failed to get submission", zap.Any("UUID", uuid), zap.Error(err))
		return nil, err
	}
	return &submission, nil
}

func NewSubmissionDataAccessor(db *mongo.Collection, logger *zap.Logger) (SubmissionDataAccessor, error) {
	return &submissionDataAccessor{db: db, logger: logger}, nil
}
