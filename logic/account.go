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

type account struct {
	logger              *zap.Logger
	accountDataAccessor db.AccountDataAccessor
	tokenLogic          Token
}

type Account interface {
	GetAccountByUUID(ctx context.Context, in *models.GetAccountRequest) (*models.GetAccountResponse, error)
	CreateAccount(ctx context.Context, in *models.CreateAccountRequest) (*models.CreateAccountResponse, error)
	GetAllAccounts(ctx context.Context) (*models.GetAccountListResponse, error)
	UpdateAccount(ctx context.Context, in *models.UpdateAccountRequest) (*models.UpdateAccountResponse, error)
	DeleteAccount(ctx context.Context, in *models.DeleteAccountRequest) error
	CreateSession(ctx context.Context, in *models.CreateSessionRequest) (*models.CreateSessionResponse, error)
	GetAccountByUsername(ctx context.Context, username string) (*models.GetAccountResponse, error)
}

func (a *account) GetAccountByUsername(ctx context.Context, username string) (*models.GetAccountResponse, error) {
	account, err := a.accountDataAccessor.GetAccountByUsername(ctx, username)
	if err != nil {
		a.logger.Error("failed to get account by username", zap.Error(err), zap.String("username", username))
		return nil, err
	}
	return &models.GetAccountResponse{Account: *account}, nil
}

func (a *account) GetAccountByUUID(ctx context.Context, in *models.GetAccountRequest) (*models.GetAccountResponse, error) {
	account, err := a.accountDataAccessor.GetAccountByUUID(ctx, in.UUID)
	if err != nil {
		a.logger.Error("fail to get account by uuid", zap.Error(err))
		return nil, err
	}
	return &models.GetAccountResponse{Account: *account}, nil
}

func (a *account) CreateAccount(ctx context.Context, in *models.CreateAccountRequest) (*models.CreateAccountResponse, error) {
	currentTime := utils.FormatTime(time.Now())

	account := db.Account{
		UUID:      uuid.NewString(),
		Username:  in.Username,
		Password:  in.Password,
		Role:      in.Role,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	err := a.accountDataAccessor.CreateAccount(ctx, &account)
	if err != nil {
		a.logger.Error("fail to add an account into database", zap.Any("account", account))
		return nil, err
	}

	a.logger.Info("Successfully created an account", zap.Any("account", account))
	return &models.CreateAccountResponse{
		Username: account.Username,
		Role:     account.Role,
	}, nil
}

func (a *account) GetAllAccounts(ctx context.Context) (*models.GetAccountListResponse, error) {
	listOfAccounts, err := a.accountDataAccessor.GetAllAccounts(ctx)
	if err != nil {
		return &models.GetAccountListResponse{}, err
	}
	totalCount := len(*listOfAccounts)
	response := &models.GetAccountListResponse{
		ListOfAccounts: *listOfAccounts,
		TotalCount:     totalCount,
	}
	a.logger.Info("successfully retrieved all accounts", zap.Int("count", totalCount))
	return response, nil
}

func (a *account) UpdateAccount(ctx context.Context, in *models.UpdateAccountRequest) (*models.UpdateAccountResponse, error) {
	currentTime := utils.FormatTime(time.Now())

	update := bson.M{
		"$set": bson.M{
			"role":      in.RequestingRole,
			"updatedAt": currentTime,
		},
	}

	err := a.accountDataAccessor.UpdateAccount(ctx, in.UUID, update)
	if err != nil {
		a.logger.Error("fail to update account", zap.Error(err))
		return &models.UpdateAccountResponse{}, err
	}

	return &models.UpdateAccountResponse{
		UUID:      in.UUID,
		Role:      in.RequestingRole,
		UpdatedAt: currentTime,
	}, nil
}

func (a *account) DeleteAccount(ctx context.Context, in *models.DeleteAccountRequest) error {
	err := a.accountDataAccessor.DeleteAccount(ctx, in.UUID)
	if err != nil {
		a.logger.Error("fail to delete account", zap.Error(err))
		return err
	}
	a.logger.Info("Successfully deleted account", zap.String("UUID", in.UUID))
	return nil
}

// CreaateSession implements Account.
func (a *account) CreateSession(ctx context.Context, in *models.CreateSessionRequest) (*models.CreateSessionResponse, error) {
	account, err := a.accountDataAccessor.GetAccountByUsername(ctx, in.Username)
	if err != nil {
		a.logger.Error("failed to get account by username", zap.Error(err), zap.String("username", in.Username))
		return &models.CreateSessionResponse{}, err
	}

	if account.Password != in.Password {
		a.logger.Info("Incorrect Password")
		return &models.CreateSessionResponse{}, fmt.Errorf("incorrect password")
	}
	token, _, err := a.tokenLogic.GetToken(ctx, account.Username, account.Role)
	if err != nil {
		a.logger.Error("failed to create token", zap.Error(err), zap.String("username", in.Username))
		return &models.CreateSessionResponse{}, err
	}

	return &models.CreateSessionResponse{
		Token:       token,
		Username:    account.Username,
		Role:        account.Role,
		AccountUUID: account.UUID,
	}, nil
}

func NewAccountLogic(logger *zap.Logger, accountDataAccessor db.AccountDataAccessor, token Token) Account {
	return &account{
		logger:              logger,
		accountDataAccessor: accountDataAccessor,
		tokenLogic:          token,
	}
}
