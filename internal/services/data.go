package services

import "sync"

// Балансы пользователей
var balances = map[string]int64{
	"user": 1000,
}

// Мьютекс для синхронизации доступа к балансам
var balancesMutex = &sync.Mutex{}
