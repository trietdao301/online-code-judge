package logic

import (
	"context"
	"strings"
	"time"

	"example/server/handlers/models"

	"example/server/db"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Submission interface {
	GetSubmission(ctx context.Context, in *models.GetSubmissionRequest) (*models.GetSubmissionResponse, error)
	CreateSubmission(ctx context.Context, in *models.CreateSubmissionRequest) (*models.CreateSubmissionResponse, error)
	DeleteSubmission(ctx context.Context, in *models.DeleteSubmissionRequest) error
	UpdateSubmission(ctx context.Context, in *models.UpdateSubmissionRequest) error
	GetSubmissionsByProblemAndAuthor(ctx context.Context, problemUUID, authorAccountUUID string) (*models.GetSubmissionListResponse, error)
}

type submission struct {
	db                     *mongo.Client
	logger                 *zap.Logger
	judge                  Judge
	submissionDataAccessor db.SubmissionDataAccessor
}

// CreateSubmission implements Submission.
func (s *submission) GetSubmission(ctx context.Context, in *models.GetSubmissionRequest) (*models.GetSubmissionResponse, error) {
	submissionRes, err := s.submissionDataAccessor.GetSubmissionByUUID(ctx, in.UUID)
	if err != nil {
		s.logger.Error("fail to create submission database", zap.Error(err))
		return &models.GetSubmissionResponse{}, err
	}
	return &models.GetSubmissionResponse{Submission: *submissionRes}, nil
}

// CreateSubmission implements Submission.
func (s *submission) CreateSubmission(ctx context.Context, in *models.CreateSubmissionRequest) (*models.CreateSubmissionResponse, error) {
	s.logger.Info("Creating Submission...", zap.Any("authorID: ", in.AuthorAccountUUID))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var UUID = uuid.NewString()
	submission := &db.Submission{
		UUID:              UUID,
		ProblemUUID:       in.ProblemUUID,
		AuthorAccountUUID: in.AuthorAccountUUID,
		Content:           in.Content,
		Language:          strings.ToLower(in.Language),
		Status:            db.SubmissionStatusSubmitted,
		GradingResult:     "",
		CreatedTime:       time.Now().UnixMilli(),
	}
	err := s.submissionDataAccessor.CreateSubmission(ctx, submission)
	if err != nil {
		return nil, err
	}
	time.Sleep(1 * time.Second)
	s.judge.ScheduleJudgeLocalSubmission(UUID)
	return &models.CreateSubmissionResponse{Submission: *submission}, nil
}

func (s *submission) GetSubmissionsByProblemAndAuthor(ctx context.Context, problemUUID, authorAccountUUID string) (*models.GetSubmissionListResponse, error) {
	s.logger.Info("Getting submissions by problem and author",
		zap.String("problemUUID", problemUUID),
		zap.String("authorAccountUUID", authorAccountUUID))
	var response models.GetSubmissionListResponse

	submissions, err := s.submissionDataAccessor.GetSubmissionsByProblemAndAuthor(ctx, problemUUID, authorAccountUUID)
	if err != nil {
		s.logger.Error("Failed to get submissions by problem and author",
			zap.String("problemUUID", problemUUID),
			zap.String("authorAccountUUID", authorAccountUUID),
			zap.Error(err))
		return nil, err
	}
	response.Submissions = submissions
	return &response, nil
}

// CreateSubmission implements Submission.
func (s *submission) UpdateSubmission(ctx context.Context, in *models.UpdateSubmissionRequest) error {
	panic("unimplemented")
}

// CreateSubmission implements Submission.
func (s *submission) DeleteSubmission(ctx context.Context, in *models.DeleteSubmissionRequest) error {
	panic("unimplemented")
}

func NewSubmissionLogic(j Judge, logger *zap.Logger, client *mongo.Client, submissionDataAccessor db.SubmissionDataAccessor) (s Submission) {
	return &submission{db: client, judge: j, logger: logger, submissionDataAccessor: submissionDataAccessor}
}
