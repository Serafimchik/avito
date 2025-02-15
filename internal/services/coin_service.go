package services

import "errors"

// Логика отправки монет
func SendCoins(fromUser, toUser string, amount int64) error {
	balancesMutex.Lock()
	defer balancesMutex.Unlock()

	if balance, ok := balances[fromUser]; ok && balance >= amount {
		balances[fromUser] -= amount
		balances[toUser] += amount
		return nil
	}
	return errors.New("insufficient funds or invalid user")
}
