package usecases

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/onmono/internal/balance"
	"github.com/onmono/internal/balance/converter"
	"github.com/onmono/internal/balance/models"
	"github.com/onmono/pkg/logging"
	"github.com/pkg/errors"
	"time"
)

type UseCase struct {
	ctx    context.Context
	repo   balance.Repository
	logger *logging.Logger
}

func NewUseCase(ctx context.Context, repo balance.Repository, logger *logging.Logger) *UseCase {
	return &UseCase{
		ctx, repo, logger,
	}
}

type DepositDTO struct {
	ID      uuid.UUID `json:"id"`
	Deposit float64   `json:"deposit"`
}

type DebitingDTO struct {
	ID    uuid.UUID `json:"id"`
	Debit float64   `json:"debit"`
}

type TransferDTO struct {
	FromId uuid.UUID `json:"from_id"`
	ToId   uuid.UUID `json:"to_id"`
	Money  float64   `json:"money"`
}

func (uc *UseCase) GetBalance(ctx context.Context, dto models.UserBalance) (model models.UserBalance, err error) {
	dto, err = uc.repo.FindOne(ctx, dto.UserID)
	if err != nil {
		uc.logger.Error(err)
		return models.UserBalance{}, err
	}
	return dto, nil
}

func (uc *UseCase) Create(ctx context.Context, dto models.UserBalance) (model models.UserBalance, err error) {
	connTx, err := uc.repo.Create(ctx, dto)
	defer connTx.Conn.Release()

	if err != nil {
		uc.logger.Error(err)
		return models.UserBalance{}, err
	}

	if err != nil {
		err = connTx.Tx.Rollback(ctx)
		if err != nil {
			return models.UserBalance{}, err
		}
		return models.UserBalance{}, err
	}
	return dto, nil
}

func (uc *UseCase) Deposit(ctx context.Context, dto DepositDTO) (models.UserBalance, error) {
	dbModel, err := uc.repo.FindOne(ctx, dto.ID)
	if err != nil && err.Error() == "no rows in result set" {
		model := models.UserBalance{
			UserID:  dto.ID,
			Balance: converter.ReduceDenomination(dto.Deposit),
		}
		connTx, err := uc.repo.Create(ctx, model)
		defer connTx.Conn.Release()
		if err != nil {
			return models.UserBalance{}, err
		}
		err = connTx.Tx.Commit(ctx)
		if err != nil {
			err = connTx.Tx.Rollback(ctx)
			if err != nil {
				return models.UserBalance{}, err
			}
			return models.UserBalance{}, err
		}
		return model, nil
	}
	if dto.Deposit >= 0 {
		dbModel.Balance = dbModel.Balance + converter.ReduceDenomination(dto.Deposit)
	}
	connTx, err := uc.repo.Update(context.TODO(), dbModel)
	defer connTx.Conn.Release()
	if err != nil {
		uc.logger.Error(err)
		return models.UserBalance{}, err
	}
	err = connTx.Tx.Commit(ctx)
	if err != nil {
		err := connTx.Tx.Rollback(ctx)
		if err != nil {
			return models.UserBalance{}, err
		}
		return models.UserBalance{}, err
	}
	return dbModel, nil
}

func (uc *UseCase) Revenue(ctx context.Context, dto models.Reserve) (models.AccountingRevenue, error) {
	reserves, err := uc.repo.GetReserve(ctx, dto)
	if err != nil {
		return models.AccountingRevenue{}, err
	}
	if len(reserves) == 0 {
		return models.AccountingRevenue{}, fmt.Errorf("no revenue to created")
	}
	reserve := reserves[0]

	debFirstBalance, err := uc.Debiting(ctx, DebitingDTO{
		ID:    reserve.UserID,
		Debit: converter.Convert(converter.Currency(reserves[0].Price)),
	})
	result := models.AccountingRevenue{
		ID:        uuid.New(),
		UserID:    reserve.UserID,
		ServiceID: reserve.ServiceID,
		OrderID:   reserve.OrderID,
		Sum:       reserve.Price,
		Timestamp: time.Now().UTC(),
	}
	if err != nil {
		uc.logger.Printf("revenue debiting user balance %v cancel with error %v", debFirstBalance, err)
		return result, err
	}

	//reserves = append(reserves[:0], reserves[1:]...)

	for _, v := range reserves {
		err := uc.DeleteReserve(ctx, v)
		if err != nil {
			return models.AccountingRevenue{}, err
		}
		err = uc.DeleteBalance(ctx, v.ReserveID)
		if err != nil {
			uc.logger.Errorf("error delete balance, %v", v.ReserveID)
		}
	}

	// добавить в отчет accounting_revenue reserves[0]
	connTx, err := uc.repo.CreateRevenue(ctx, result)
	if err != nil {
		connTx.Conn.Release()
		return models.AccountingRevenue{}, err
	}
	connTx.Tx.Commit(ctx)
	return result, err
}

func (uc *UseCase) Reserve(ctx context.Context, dto models.Reserve) (models.Reserve, error) {
	model, err := uc.GetBalance(ctx, models.UserBalance{UserID: dto.UserID})
	if err != nil {
		return models.Reserve{}, errors.New("no user balance with current user_id for reserve")
	}
	// require price > 0 and balance >= price
	if dto.Price < 0 || model.Balance < dto.Price {
		return models.Reserve{}, errors.New("require price greatest than 0 and user balance greatest than price")
	}

	reservedUser := models.UserBalance{
		ID:            uuid.New(),
		UserID:        uuid.New(),
		Balance:       dto.Price,
		LastUpdatedAt: time.Now().UTC(),
	}

	connTxCreate, err := uc.repo.Create(ctx, reservedUser)
	defer connTxCreate.Conn.Release()
	if err != nil {
		return models.Reserve{}, errors.New("reserve user balance not created")
	}

	connTxCreate.Tx.Commit(ctx)
	defer connTxCreate.Conn.Release()

	reserve := models.Reserve{
		ID:            uuid.New(),
		ReserveID:     reservedUser.UserID,
		UserID:        dto.UserID,
		ServiceID:     dto.ServiceID,
		OrderID:       dto.OrderID,
		Price:         dto.Price,
		LastUpdatedAt: time.Now(),
	}

	txReserve, err := uc.repo.Reserve(ctx, reserve)
	if err != nil {
		connTxCreate.Tx.Rollback(ctx)
		return models.Reserve{}, errors.New("reserve_info not created")
	}

	err = txReserve.Tx.Commit(ctx)
	defer txReserve.Conn.Release()

	if err != nil {
		connTxCreate.Tx.Rollback(ctx)
		return models.Reserve{}, err
	}

	return reserve, nil
}

func (uc *UseCase) DeleteReserve(ctx context.Context, dto models.Reserve) error {
	err := uc.repo.DeleteReserve(ctx, dto.ID)
	if err != nil {
		uc.logger.Error(err)
		return err
	}
	return nil
}

func (uc *UseCase) DeleteBalance(ctx context.Context, id uuid.UUID) error {
	return uc.repo.DeleteUserBalance(ctx, id)
}

func (uc *UseCase) Debiting(ctx context.Context, dto DebitingDTO) (models.UserBalance, error) {
	dbModel, err := uc.repo.FindOne(ctx, dto.ID)
	if err != nil {
		uc.logger.Error(err)
		return models.UserBalance{}, err
	}
	if dto.Debit > 0 {
		temp := dbModel.Balance - converter.ReduceDenomination(dto.Debit)
		if temp >= 0 {
			dbModel.Balance = temp
		} else {
			errMessage := "the balance should not be negative, please try again with a different amount"
			uc.logger.Error(errMessage)
			return models.UserBalance{}, fmt.Errorf(errMessage)
		}
	} else {
		errMessage := "debit should not be zero or negative"
		uc.logger.Error(errMessage)
		return models.UserBalance{}, fmt.Errorf(errMessage)
	}
	updateTx, err := uc.repo.Update(ctx, dbModel)
	defer updateTx.Conn.Release()
	if err != nil {
		uc.logger.Error(err)
		return models.UserBalance{}, err
	}
	err = updateTx.Tx.Commit(ctx)
	if err != nil {
		updateTx.Tx.Rollback(ctx)
		return models.UserBalance{}, err
	}
	return dbModel, nil
}

func (uc *UseCase) Transfer(ctx context.Context, dto TransferDTO) (err error) {
	from := DebitingDTO{
		ID:    dto.FromId,
		Debit: dto.Money,
	}
	to := DepositDTO{
		ID:      dto.ToId,
		Deposit: dto.Money,
	}
	balanceModel := models.UserBalance{
		UserID:  dto.ToId,
		Balance: converter.ReduceDenomination(dto.Money),
	}

	_, err = uc.GetBalance(ctx, balanceModel)
	if err != nil {
		errMessage := "the balance you are transferring money to does not exist yet"
		uc.logger.Error(errMessage)
		return fmt.Errorf(errMessage)
	}
	_, err = uc.Debiting(ctx, from)
	if err != nil {
		uc.logger.Error(err)
		return err
	}
	_, err = uc.Deposit(ctx, to)
	if err != nil {
		uc.logger.Error(err)
		return err
	}
	return nil
}
