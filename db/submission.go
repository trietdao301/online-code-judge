package db

import (
	"context"
	"fmt"

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
	UpdateSubmissionByUUID(ctx context.Context, uuid string, update map[string]any) error
	GetSubmissionsByProblemAndAuthor(ctx context.Context, problemUUID, authorAccountUUID string) ([]*Submission, error)
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
	s.logger.Info("getting submission by UUID", zap.Any("UUID", uuid))
	var submission Submission
	err := s.db.FindOne(ctx, filter).Decode(&submission)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			s.logger.Error("fail to find submission", zap.Any("UUID", uuid))
			return &Submission{}, nil // Return nil if no document is found
		}
		s.logger.Error("fail to get submission", zap.Any("UUID", uuid), zap.Error(err))
		return &Submission{}, err
	}
	return &submission, nil
}

func (s *submissionDataAccessor) UpdateSubmissionByUUID(ctx context.Context, uuid string, update map[string]any) error {

	updateAsBson := bson.M{
		"$set": update,
	}
	filter := bson.M{"UUID": uuid}
	result, err := s.db.UpdateOne(ctx, filter, updateAsBson)
	if err != nil {
		s.logger.Error("fail to update submission data", zap.Any("submission UUID", uuid))
		return err
	}
	if result.MatchedCount == 0 {
		s.logger.Error("fail to update submission data", zap.Any("submission UUID", uuid))
		return fmt.Errorf("no submission found with UUID: %s", uuid)
	}
	return nil
}

func (s *submissionDataAccessor) GetSubmissionsByProblemAndAuthor(ctx context.Context, problemUUID, authorAccountUUID string) ([]*Submission, error) {
	filter := bson.M{
		"problemUUID":       problemUUID,
		"authorAccountUUID": authorAccountUUID,
	}

	s.logger.Info("getting submissions by problem UUID and author account UUID",
		zap.String("problemUUID", problemUUID),
		zap.String("authorAccountUUID", authorAccountUUID))

	cursor, err := s.db.Find(ctx, filter)
	if err != nil {
		s.logger.Error("fail to find submissions",
			zap.String("problemUUID", problemUUID),
			zap.String("authorAccountUUID", authorAccountUUID),
			zap.Error(err))
		return []*Submission{}, err
	}
	defer cursor.Close(ctx)

	var submissions []*Submission
	for cursor.Next(ctx) {
		var submission Submission
		if err := cursor.Decode(&submission); err != nil {
			s.logger.Error("fail to decode submission", zap.Error(err))
			return []*Submission{}, err
		}
		submissions = append(submissions, &submission)
	}

	if err := cursor.Err(); err != nil {
		s.logger.Error("cursor error", zap.Error(err))
		return []*Submission{}, err
	}

	return submissions, nil
}

func NewSubmissionDataAccessor(db *mongo.Collection, logger *zap.Logger) (SubmissionDataAccessor, error) {
	return &submissionDataAccessor{db: db, logger: logger}, nil
}
