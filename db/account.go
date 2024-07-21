package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type AccountDataAccessor interface {
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountByUUID(ctx context.Context, UUID string) (*Account, error)
	UpdateAccount(ctx context.Context, accountUUID string, update bson.M) error
	GetAllAccounts(ctx context.Context) (*[]Account, error)
	DeleteAccount(ctx context.Context, accountUUID string) error
	GetAccountByUsername(ctx context.Context, username string) (*Account, error)
}

type accountDataAccessor struct {
	db     *mongo.Collection
	logger *zap.Logger
}

type Account struct {
	UUID      string `json:"UUID" bson:"UUID" validate:"required"`
	Username  string `json:"username" bson:"username" validate:"required"`
	Password  string `json:"password" bson:"password" validate:"required"`
	Role      string `json:"role" bson:"role" validate:"oneof=Admin Contestant ProblemSetter"`
	CreatedAt string `json:"createdAt" bson:"createdAt"`
	UpdatedAt string `json:"updatedAt" bson:"updatedAt"`
}

func (a *accountDataAccessor) CreateAccount(ctx context.Context, account *Account) error {
	filter := bson.M{
		"username": account.Username,
	}
	var usedAccount Account
	err := a.db.FindOne(ctx, filter).Decode(&usedAccount)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			a.logger.Info("Creating account")
			_, err = a.db.InsertOne(ctx, account)
			if err != nil {
				a.logger.Error("fail to create account in database", zap.Error(err))
				return err
			}
			return nil
		}
	}
	a.logger.Error("Account is already existed")
	return fmt.Errorf("Account is already existed")
}

func (a *accountDataAccessor) GetAccountByUsername(ctx context.Context, username string) (*Account, error) {
	filter := bson.M{"username": username}
	var account Account
	err := a.db.FindOne(ctx, filter).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			a.logger.Warn("no account found with the given username", zap.String("username", username))
			return nil, fmt.Errorf("no account found with username: %s", username)
		}
		a.logger.Error("failed to find account in database", zap.String("username", username), zap.Error(err))
		return nil, err
	}
	a.logger.Info("successfully retrieved account by username", zap.String("username", username))
	return &account, nil
}

func (a *accountDataAccessor) GetAccountByUUID(ctx context.Context, UUID string) (*Account, error) {
	filter := bson.M{"UUID": UUID}
	var account Account
	err := a.db.FindOne(ctx, filter).Decode(&account)
	if err != nil {
		a.logger.Error("fail to find account in database", zap.Error(err))
		return nil, err
	}
	return &account, nil
}

func (a *accountDataAccessor) UpdateAccount(ctx context.Context, accountUUID string, update bson.M) error {
	filter := bson.M{"UUID": accountUUID}
	result, err := a.db.UpdateOne(ctx, filter, update)
	if err != nil {
		a.logger.Error("failed to update account", zap.String("accountUUID", accountUUID), zap.Error(err))
		return err
	}
	if result.MatchedCount == 0 {
		a.logger.Warn("no account found to update", zap.String("accountUUID", accountUUID))
		return fmt.Errorf("no account found with UUID: %s", accountUUID)
	}
	a.logger.Info("account updated successfully", zap.String("accountUUID", accountUUID))
	return nil
}

func (a *accountDataAccessor) GetAllAccounts(ctx context.Context) (*[]Account, error) {
	var accounts []Account
	cursor, err := a.db.Find(ctx, bson.M{})
	if err != nil {
		a.logger.Error("failed to create cursor for all accounts", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var account Account
		if err := cursor.Decode(&account); err != nil {
			a.logger.Error("failed to decode account", zap.Error(err))
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err := cursor.Err(); err != nil {
		a.logger.Error("cursor encountered an error", zap.Error(err))
		return nil, err
	}

	a.logger.Info("successfully retrieved all accounts", zap.Int("count", len(accounts)))
	return &accounts, nil
}

func (a *accountDataAccessor) DeleteAccount(ctx context.Context, accountUUID string) error {
	filter := bson.M{"UUID": accountUUID}
	_, err := a.db.DeleteOne(ctx, filter)
	if err != nil {
		a.logger.Error("fail to delete account", zap.String("accountUUID", accountUUID), zap.Error(err))
		return err
	}
	return nil
}

func NewAccountDataAccessor(db *mongo.Collection, logger *zap.Logger) (AccountDataAccessor, error) {
	return &accountDataAccessor{db: db, logger: logger}, nil
}
