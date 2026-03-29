package yookassa

import (
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/rvinnie/yookassa-sdk-go/yookassa"
	yoocommon "github.com/rvinnie/yookassa-sdk-go/yookassa/common"
	yoopayment "github.com/rvinnie/yookassa-sdk-go/yookassa/payment"
)

type YookassaPayment struct {
	paymentHandler *yookassa.PaymentHandler
	redirectUrl    string
}

func New(paymentHandler *yookassa.PaymentHandler, redirectUrl string) *YookassaPayment {
	return &YookassaPayment{
		paymentHandler: paymentHandler,
		redirectUrl:    redirectUrl,
	}
}

func (y *YookassaPayment) CreatePayment(userId int64, price float32) (string, error) {
	const op = "lib.payment.yookassa.CreatePayment"

	payment, err := y.paymentHandler.CreatePayment(&yoopayment.Payment{
		Amount: &yoocommon.Amount{
			Value:    fmt.Sprint(price),
			Currency: "RUB",
		},
		PaymentMethod: yoopayment.PaymentTypeBankCard,
		Confirmation: yoopayment.Redirect{
			Type:      "redirect",
			ReturnURL: y.redirectUrl,
		},
		Description: "Оплата в магазине 3",
		Metadata: map[string]any{
			"user_id": userId, //! this may make bugs, if user creates 2 orders in short time delta
		},
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, errs.ErrCreatePayment)
	}

	return payment.Confirmation.(map[string]any)["confirmation_url"].(string), nil
}
