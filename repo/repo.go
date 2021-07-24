package repo

import (
	"debts_bot/pkg"
	"gorm.io/gorm"
	"time"
)

type Debts interface {
	Save(debt *pkg.Debt) error
	GetListByLenderID(id uint) ([]*pkg.Debt, error)
	GetListByDebtorID(id uint) ([]*pkg.Debt, error)
	GetActiveListByLenderID(id uint) ([]*pkg.Debt, error)
	GetActiveListByDebtorID(id uint) ([]*pkg.Debt, error)
	GetByDebtID(id uint) (*pkg.Debt, error)
	SetStatus(id uint, status string) error
}

func (d *debts) GetActiveListByLenderID(id uint) (ans []*pkg.Debt, err error) {
	err = d.db.Where("lender_id = ?", id).Where("status = ? or status = ?", pkg.DebtStatusActive, pkg.DebtStatusStopWaiting).Order("sum").Find(&ans).Error
	return
}

func (d *debts) GetActiveListByDebtorID(id uint) (ans []*pkg.Debt, err error) {
	err = d.db.Where("debtor_id = ?", id).Where("status = ? or status = ?", pkg.DebtStatusActive, pkg.DebtStatusStopWaiting).Order("sum").Find(&ans).Error
	return
}

type debts struct {
	db *gorm.DB
}

func (d *debts) Save(debt *pkg.Debt) error {
	return d.db.Save(debt).Error
}

func (d *debts) GetListByLenderID(id uint) (ans []*pkg.Debt, err error) {
	err = d.db.Where("lender_id = ?", id).Find(&ans).Error
	return
}

func (d *debts) GetListByDebtorID(id uint) (ans []*pkg.Debt, err error) {
	err = d.db.Where("debtor_id = ?", id).Find(&ans).Error
	return
}

func (d *debts) GetByDebtID(id uint) (*pkg.Debt, error) {
	var ans pkg.Debt
	err := d.db.Where("id = ?", id).First(&ans).Error
	return &ans, err
}

func (d *debts) SetStatus(id uint, status string) error {
	debt, err := d.GetByDebtID(id)
	if err != nil {
		return err
	}

	debt.Status = status
	if status == pkg.DebtStatusClosed {
		debt.ClosedAt = time.Now()
	}

	return d.db.Save(debt).Error
}

func NewDebts(db *gorm.DB) (Debts, error) {
	err := db.AutoMigrate(&pkg.Debt{})

	return &debts{db: db}, err
}
