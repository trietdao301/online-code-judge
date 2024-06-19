package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type TestCase struct {
	UUID            string `json:"UUID" bson:"UUID"`
	OfProblemUUID   string `json:"ofProblemUUID" bson:"ofProblemUUID"`
	TestFileContent string `json:"TestFileContent" bson:"TestFileContent"`
}

type TestCaseDataAccessor interface {
	CreateTestCase(ctx context.Context, testCase *TestCase) error
	GetTestCaseByProblemUUID(ctx context.Context, UUID string) (*TestCase, error)
	GetTestCaseByUUID(ctx context.Context, UUID string) (*TestCase, error)
}

type testCaseDataAccessor struct {
	db     *mongo.Collection
	logger *zap.Logger
}

// CreateTestCase implements TestCaseDataAccesor.
func (t testCaseDataAccessor) CreateTestCase(ctx context.Context, testCase *TestCase) error {
	_, err := t.db.InsertOne(ctx, testCase)
	t.logger.Info(t.db.Database().Name())
	if err != nil {
		t.logger.Error(err.Error())
		return err
	}
	return nil
}

// GetTestCase implements TestCaseDataAccesor.
func (t testCaseDataAccessor) GetTestCaseByProblemUUID(ctx context.Context, problemUUID string) (*TestCase, error) {
	filter := bson.M{"ofProblemUUID": problemUUID}
	var testCase TestCase
	err := t.db.FindOne(ctx, filter).Decode(&testCase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		t.logger.Error("Failed to get test case", zap.String("uuid", problemUUID), zap.Error(err))
		return nil, err
	}
	return &testCase, nil
}

// GetTestCaseByUUID retrieves a test case by its UUID.
func (t testCaseDataAccessor) GetTestCaseByUUID(ctx context.Context, uuid string) (*TestCase, error) {
	filter := bson.M{"UUID": uuid}
	var testCase TestCase
	err := t.db.FindOne(ctx, filter).Decode(&testCase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Return nil if no document is found
		}
		t.logger.Error("Failed to get test case by UUID", zap.String("UUID", uuid), zap.Error(err))
		return nil, err
	}
	return &testCase, nil
}

func NewTestCaseDataAccessor(db *mongo.Collection, logger *zap.Logger) (TestCaseDataAccessor, error) {
	return &testCaseDataAccessor{db: db, logger: logger}, nil
}
