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
	TestFileContent string `json:"testFileContent" bson:"testFileContent"`
	CreatedAt       string `json:"createdAt" bson:"createdAt"`
	Language        string `json:"language" bson:"language"`
}

type TestCaseDataAccessor interface {
	CreateTestCase(ctx context.Context, testCase *TestCase) error
	GetTestCaseByProblemUUID(ctx context.Context, UUID string) (*TestCase, error)
	GetTestCaseByUUID(ctx context.Context, UUID string) (*TestCase, error)
	DeleteTestCase(ctx context.Context, testCaseUUID string) error
	GetTestCaseByProblemUUIDAndLanguage(ctx context.Context, problemUUID string, language string) (*TestCase, error)
}

type testCaseDataAccessor struct {
	db     *mongo.Collection
	logger *zap.Logger
}

func (t testCaseDataAccessor) GetTestCaseByProblemUUIDAndLanguage(ctx context.Context, problemUUID string, language string) (*TestCase, error) {
	filter := bson.M{
		"ofProblemUUID": problemUUID,
		"language":      language,
	}
	var testCase TestCase
	err := t.db.FindOne(ctx, filter).Decode(&testCase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			t.logger.Info("No test case found", zap.String("problemUUID", problemUUID), zap.String("language", language))
			return nil, nil // Return nil if no document is found
		}
		t.logger.Error("Failed to get test case", zap.String("problemUUID", problemUUID), zap.String("language", language), zap.Error(err))
		return nil, err
	}
	t.logger.Info("Retrieved test case", zap.String("UUID", testCase.UUID), zap.String("problemUUID", problemUUID), zap.String("language", language))
	return &testCase, nil
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
			return &TestCase{}, nil
		}
		t.logger.Error("Failed to get test case", zap.String("uuid", problemUUID), zap.Error(err))
		return &TestCase{}, err
	}
	return &testCase, nil
}

// GetTestCaseByUUID retrieves a test case by its UUID.
func (t testCaseDataAccessor) GetTestCaseByUUID(ctx context.Context, UUID string) (*TestCase, error) {
	filter := bson.M{"UUID": UUID}
	var testCase TestCase
	err := t.db.FindOne(ctx, filter).Decode(&testCase)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			t.logger.Error("Failed to find test case with UUID", zap.String("UUID", UUID), zap.Error(err))
			return &TestCase{}, nil // Return nil if no document is found
		}
		t.logger.Error("Failed to get test case by UUID", zap.String("UUID", UUID), zap.Error(err))
		return &TestCase{}, nil
	}
	t.logger.Info("get test UUID here: ", zap.Any("test: ", testCase))
	return &testCase, nil
}

func (t testCaseDataAccessor) DeleteTestCase(ctx context.Context, testCaseUUID string) error {
	filter := bson.M{
		"UUID": testCaseUUID,
	}
	_, err := t.db.DeleteOne(ctx, filter)
	if err != nil {
		t.logger.Error("fail to delete testCase", zap.Any("testCaseUUID", testCaseUUID), zap.Any("error: ", err.Error()))
		return err
	}
	return nil
}

func NewTestCaseDataAccessor(db *mongo.Collection, logger *zap.Logger) (TestCaseDataAccessor, error) {
	return &testCaseDataAccessor{db: db, logger: logger}, nil
}
