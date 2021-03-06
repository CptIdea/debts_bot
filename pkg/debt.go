package pkg

import (
	"gorm.io/gorm"
	"time"
)

type Debt struct {
	gorm.Model
	LenderID    int64
	DebtorID    int64
	AuthorID    int64
	Status      string `gorm:"default:ожидание начала"`
	Sum         int64
	Currency    string `gorm:"default:₽"`
	Description string

	LastNotify time.Time
	ClosedAt   time.Time
}

var (
	DebtStatusStartWaiting = "ожидание начала"
	DebtStatusActive       = "в процессе"
	DebtStatusClosed       = "закрыт"
	DebtStatusStopWaiting  = "ожидание закрытия"
	DebtStatusCanceled     = "отменен"
)
