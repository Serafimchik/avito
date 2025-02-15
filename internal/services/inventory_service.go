package services

import "errors"

// Список доступных товаров и их цены
var items = map[string]int64{
	"t-shirt":    80,
	"cup":        20,
	"book":       50,
	"pen":        10,
	"powerbank":  200,
	"hoody":      300,
	"umbrella":   200,
	"socks":      10,
	"wallet":     50,
	"pink-hoody": 500,
}

// Инвентарь пользователей
var inventory = map[string][]InventoryItem{}

func GetUserInfo(username string) (map[string]interface{}, error) {
	balancesMutex.Lock()
	defer balancesMutex.Unlock()

	if _, exists := balances[username]; !exists {
		return nil, errors.New("user not found")
	}

	userInfo := map[string]interface{}{
		"coins":       balances[username],
		"inventory":   inventory[username],
		"coinHistory": getCoinHistory(username),
	}
	return userInfo, nil
}

func getCoinHistory(username string) map[string][]CoinTransaction {
	return map[string][]CoinTransaction{
		"received": {},
		"sent":     {},
	}
}

// Логика покупки предмета
func BuyItem(username, item string) error {
	balancesMutex.Lock()
	defer balancesMutex.Unlock()

	if price, ok := items[item]; ok {
		if balance, exists := balances[username]; exists && balance >= price {
			balances[username] -= price
			inventory[username] = append(inventory[username], InventoryItem{Type: item, Quantity: 1})
			return nil
		}
	}
	return errors.New("insufficient funds or invalid item")
}

// Модели данных
type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int64  `json:"quantity"`
}
type CoinTransaction struct {
	User   string `json:"user"`
	Amount int64  `json:"amount"`
}
