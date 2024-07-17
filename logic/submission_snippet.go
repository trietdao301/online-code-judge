package logic

import (
	"context"
	"example/server/db"
	"example/server/handlers/models"
	"example/server/utils"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type SubmissionSnippet interface {
	GetSubmissionSnippetByUUID(ctx context.Context, in *models.GetSubmissionSnippetRequest) (*models.GetSubmissionSnippetResponse, error)
	CreateSubmissionSnippet(ctx context.Context, in *models.CreateSubmissionSnippetRequest) error
}
type submissionSnippet struct {
	logger                        *zap.Logger
	submissionSnippetDataAccessor db.SubmissionSnippetDataAccessor
	problemDataAccessor           db.ProblemDataAccessor
}

func (s submissionSnippet) GetSubmissionSnippetByUUID(ctx context.Context, in *models.GetSubmissionSnippetRequest) (*models.GetSubmissionSnippetResponse, error) {
	submissionSnippet, err := s.submissionSnippetDataAccessor.GetSubmissionSnippetByUUID(ctx, in.SubmissionSnippetUUID)
	if err != nil {
		s.logger.Error("fails during submission snippet data requesting", zap.Any("uuid", in.SubmissionSnippetUUID))
		return &models.GetSubmissionSnippetResponse{}, err
	}
	return &models.GetSubmissionSnippetResponse{CodeSnippet: submissionSnippet.CodeSnippet, Language: submissionSnippet.Language}, nil
}

func (s submissionSnippet) CreateSubmissionSnippet(ctx context.Context, in *models.CreateSubmissionSnippetRequest) error {
	var (
		submissionSnippetToAdd *db.SubmissionSnippetData
		problemToUpdate        *db.Problem
	)
	newSubmissionSnippetUUID := uuid.NewString()
	SubmissionSnippetListFieldName := utils.GetFieldName(db.Problem{}, "SubmissionSnippetList")
	submissionSnippetToAdd = &db.SubmissionSnippetData{
		SubmissionSnippetUUID: newSubmissionSnippetUUID,
		Language:              in.Language,
	}
	update := bson.M{
		"$push": bson.M{
			SubmissionSnippetListFieldName: submissionSnippetToAdd,
		},
	}
	problemToUpdate, err := s.problemDataAccessor.GetProblemByUUID(ctx, in.OfProblemUUID)
	if err != nil {
		s.logger.Error("fail to get problem by uuid")
	}
	for _, submissionSnippet := range problemToUpdate.SubmissionSnippetList {
		if submissionSnippet.Language == in.Language {
			s.logger.Info("submission snippet for language already exists, please remove it before adding a new one",
				zap.String("language", submissionSnippet.Language))
			return fmt.Errorf("test case for language %s already exists, please remove it before adding a new one", submissionSnippet.Language)
		}
	}

	currentTime := utils.FormatTime(time.Now())
	newSubmissionSnippet := &db.SubmissionSnippet{
		UUID:          newSubmissionSnippetUUID,
		OfProblemUUID: in.OfProblemUUID,
		CodeSnippet:   in.CodeSnippet,
		Language:      in.Language,
		CreatedAt:     currentTime,
	}
	err = s.submissionSnippetDataAccessor.CreateSubmissionSnippet(ctx, newSubmissionSnippet)
	if err != nil {
		s.logger.Error("failed to create submission snippet", zap.Error(err), zap.Any("snippet", submissionSnippetToAdd))
		return err
	}

	err = s.problemDataAccessor.UpdateProblem(ctx, in.OfProblemUUID, update)
	if err != nil {
		s.logger.Error("failed to update problem with new test case",
			zap.String("problemUUID", in.OfProblemUUID),
			zap.String("submissionSnippetUUID", newSubmissionSnippet.UUID),
			zap.Error(err))
		return err
	}

	return nil
}

func NewSubmissionSnippetLogic(logger *zap.Logger,
	submissionSnippetDataAccessor db.SubmissionSnippetDataAccessor,
	problemDataAccessor db.ProblemDataAccessor,

) SubmissionSnippet {
	return &submissionSnippet{
		logger:                        logger,
		submissionSnippetDataAccessor: submissionSnippetDataAccessor,
		problemDataAccessor:           problemDataAccessor,
	}
}
