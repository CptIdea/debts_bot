package notificator

import (
	"debts_bot/pkg"
)

type Notificator interface {
	SendNotify(debt *pkg.Debt, initiator int)
	NewStatusNotify(debt *pkg.Debt, initiator int)
	NewDebtNotify(debt *pkg.Debt, initiator int)
	ConfirmStopNotify(debt *pkg.Debt, initiator int)
	GetNameById(id int) string
	GenMessageFromDebt(debt *pkg.Debt) string
}
