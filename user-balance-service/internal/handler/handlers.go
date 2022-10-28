package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/onmono/internal/appresponse"
	"github.com/onmono/internal/balance/converter"
	"github.com/onmono/internal/balance/models"
	"github.com/onmono/internal/usecases"
	"github.com/onmono/pkg/logging"
	convert "github.com/onmono/pkg/utils"
	"net/http"
	"time"
)

type BalanceHandler struct {
	ctx     context.Context
	useCase *usecases.UseCase
	logger  *logging.Logger
}

func NewBalanceHandler(ctx context.Context, useCase *usecases.UseCase, logger *logging.Logger) *BalanceHandler {
	return &BalanceHandler{
		ctx, useCase, logger,
	}
}

type ReserveReq struct {
	ID            uuid.UUID `json:"id,omitempty"`
	ReserveID     uuid.UUID `json:"reserve_id,omitempty"`
	UserID        uuid.UUID `json:"user_id"`
	ServiceID     uuid.UUID `json:"service_id"`
	OrderID       uuid.UUID `json:"order_id"`
	Price         float64   `json:"price"`
	LastUpdatedAt time.Time `json:"last_updated_at,omitempty"`
}

type RevenueResp struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ServiceID uuid.UUID `json:"service_id"`
	OrderID   uuid.UUID `json:"order_id"`
	Sum       float64   `json:"sum"`
	Timestamp time.Time `json:"timestamp"`
}

type RevenueReq struct {
	UserID    uuid.UUID `json:"user_id"`
	ServiceID uuid.UUID `json:"service_id"`
	OrderID   uuid.UUID `json:"order_id"`
	Sum       float64   `json:"sum"`
}

func (h *BalanceHandler) Revenue(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	in := RevenueReq{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&in)

	if err != nil {
		message := appresponse.Message{
			Code:             http.StatusInternalServerError,
			Message:          err.Error(),
			DeveloperMessage: "something wrong with body parse",
		}
		h.logger.Error(message)
		w.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.Marshal(message)
		w.Write(resp)
		return
	}
	revenue, err := h.useCase.Revenue(context.Background(), models.Reserve{
		UserID:    in.UserID,
		ServiceID: in.ServiceID,
		OrderID:   in.OrderID,
		Price:     converter.ReduceDenomination(in.Sum),
	})
	if err != nil {
		message := appresponse.Message{
			Code:             http.StatusBadRequest,
			Message:          err.Error(),
			DeveloperMessage: "",
		}
		h.logger.Error(message)
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(message)
		w.Write(resp)
		return
	}

	result := RevenueResp{
		ID:        revenue.ID,
		UserID:    revenue.UserID,
		ServiceID: revenue.ServiceID,
		OrderID:   revenue.OrderID,
		Sum:       converter.Convert(converter.Currency(revenue.Sum)),
		Timestamp: revenue.Timestamp,
	}

	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(result)
	w.Write(resp)
}

func (h *BalanceHandler) Reserve(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	in := ReserveReq{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&in)

	if err != nil {
		message := appresponse.Message{
			Code:             http.StatusInternalServerError,
			Message:          err.Error(),
			DeveloperMessage: "something wrong with body parse",
		}
		h.logger.Error(message)
		w.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.Marshal(message)
		w.Write(resp)
		return
	}

	model, err := h.useCase.Reserve(context.Background(), models.Reserve{
		UserID:        in.UserID,
		ServiceID:     in.ServiceID,
		OrderID:       in.OrderID,
		Price:         converter.ReduceDenomination(in.Price),
		LastUpdatedAt: time.Now().UTC(),
	})

	if err != nil {
		message := appresponse.Message{
			Code:             http.StatusBadRequest,
			Message:          err.Error(),
			DeveloperMessage: "",
		}
		h.logger.Error(message)
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(message)
		w.Write(resp)
		return
	}
	result := ReserveReq{
		ID:            model.ID,
		ReserveID:     model.ReserveID,
		UserID:        model.UserID,
		ServiceID:     model.ServiceID,
		OrderID:       model.OrderID,
		Price:         converter.Convert(converter.Currency(model.Price)),
		LastUpdatedAt: model.LastUpdatedAt,
	}

	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(result)
	w.Write(resp)
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	model := models.UserBalance{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&model)

	if err != nil {
		message := appresponse.Message{
			Code:             http.StatusInternalServerError,
			Message:          err.Error(),
			DeveloperMessage: "something wrong with body parse",
		}
		h.logger.Error(message)
		w.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.Marshal(message)
		w.Write(resp)
		return
	}
	model, err = h.useCase.GetBalance(context.Background(), model)

	if err != nil && err.Error() == "no rows in result set" {
		message := appresponse.Message{
			Code:             http.StatusNotFound,
			Message:          "no such balance user found, try depositing money",
			DeveloperMessage: "",
		}
		h.logger.Error(message)
		w.WriteHeader(http.StatusNotFound)
		resp, _ := json.Marshal(message)
		w.Write(resp)
		return
	}

	respDTO := &appresponse.ResponseDTO{
		ID:     model.UserID,
		Amount: converter.Convert(converter.Currency(model.Balance)),
	}

	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(respDTO)
	w.Write(resp)
}

func (h *BalanceHandler) DepositOrDebitBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var data map[string]interface{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		message := appresponse.Message{
			Code:             http.StatusInternalServerError,
			Message:          err.Error(),
			DeveloperMessage: "something wrong with body parse",
		}
		h.logger.Error(message)
		w.WriteHeader(http.StatusInternalServerError)
		resp, _ := json.Marshal(message)
		w.Write(resp)
		return
	}
	if v, ok := data["deposit"]; ok {
		id, err := convert.GetUUIDFromMap(data["user_id"])

		if err != nil {
			fmt.Println(err)
		}

		deposit, err := convert.GetFloatFromMap(v)
		if err != nil {
			message := appresponse.Message{
				Code:             http.StatusInternalServerError,
				Message:          "wrong request field",
				DeveloperMessage: "something wrong with data parse from request",
			}
			h.logger.Error(message)
			w.WriteHeader(http.StatusInternalServerError)
			resp, _ := json.Marshal(message)
			w.Write(resp)
			return
		}
		if deposit <= 0 {
			message := appresponse.Message{
				Code:             http.StatusBadRequest,
				Message:          "wrong deposit",
				DeveloperMessage: "deposit don`t be negative or zero",
			}
			h.logger.Error(message)
			w.WriteHeader(http.StatusInternalServerError)
			resp, _ := json.Marshal(message)
			w.Write(resp)
			return
		}

		dto := usecases.DepositDTO{
			ID:      id,
			Deposit: deposit,
		}

		_, err = h.useCase.Deposit(context.Background(), dto)
		if err != nil {
			message := appresponse.Message{
				Code:             http.StatusInternalServerError,
				Message:          err.Error(),
				DeveloperMessage: "",
			}
			h.logger.Error(message)
			w.WriteHeader(http.StatusInternalServerError)
			resp, _ := json.Marshal(message)
			w.Write(resp)
			return
		}

		message := appresponse.Message{
			Code:             http.StatusOK,
			Message:          "deposit completed",
			DeveloperMessage: "",
		}

		w.WriteHeader(http.StatusOK)
		resp, err := json.Marshal(message)
		w.Write(resp)
	} else if _, ok := data["debit"]; ok {
		id, err := convert.GetUUIDFromMap(data["id"])
		if err != nil {
			message := appresponse.Message{
				Code:             http.StatusBadRequest,
				Message:          "no such id found, try another",
				DeveloperMessage: "",
			}
			h.logger.Info(message)
			w.WriteHeader(http.StatusBadRequest)
			resp, _ := json.Marshal(message)
			w.Write(resp)
			return
		}

		debit, err := convert.GetFloatFromMap(data["debit"])

		if err != nil {
			message := appresponse.Message{
				Code:             http.StatusInternalServerError,
				Message:          "wrong request field",
				DeveloperMessage: "something wrong with data parse from request",
			}
			h.logger.Error(message)
			w.WriteHeader(http.StatusInternalServerError)
			resp, _ := json.Marshal(message)
			w.Write(resp)
			return
		}

		dto := usecases.DebitingDTO{
			ID:    id,
			Debit: debit,
		}
		_, err = h.useCase.Debiting(context.Background(), dto)
		if err != nil {
			message := appresponse.Message{
				Code:             http.StatusInternalServerError,
				Message:          err.Error(),
				DeveloperMessage: "",
			}
			h.logger.Error(message)
			w.WriteHeader(http.StatusInternalServerError)
			resp, _ := json.Marshal(message)
			w.Write(resp)
			return
		}

		message := appresponse.Message{
			Code:             http.StatusOK,
			Message:          "debit completed",
			DeveloperMessage: "",
		}

		w.WriteHeader(http.StatusOK)
		resp, err := json.Marshal(message)
		w.Write(resp)
	}
}

func (h *BalanceHandler) TransferBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var data usecases.TransferDTO

	defer r.Body.Close()
	_ = json.NewDecoder(r.Body).Decode(&data)
	err := h.useCase.Transfer(context.Background(), data)
	if err != nil {
		message := appresponse.Message{
			Code:             http.StatusBadRequest,
			Message:          err.Error(),
			DeveloperMessage: "",
		}
		h.logger.Info(message)
		w.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(message)
		w.Write(resp)
		return
	}

	message := appresponse.Message{
		Code:             http.StatusOK,
		Message:          "transfer completed",
		DeveloperMessage: "",
	}

	w.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(message)
	w.Write(resp)
}
