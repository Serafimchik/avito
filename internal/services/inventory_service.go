package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

func BuyItem(userId int, itemType string) error {
	ctx := context.Background()

	tx, err := Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
		}
	}()

	itemQuery := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("price").
		From("items").
		Where(squirrel.Eq{"type": itemType})

	var price int64
	sqlStr, args, err := itemQuery.ToSql()
	if err != nil {
		log.Printf("Failed to build item query: %v", err)
		tx.Rollback(ctx)
		return err
	}

	err = tx.QueryRow(ctx, sqlStr, args...).Scan(&price)
	if err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("item not found")
		}
		return err
	}

	balanceQuery := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("coins").
		From("users").
		Where(squirrel.Eq{"id": userId})

	var balance int64
	sqlStr, args, err = balanceQuery.ToSql()
	if err != nil {
		log.Printf("Failed to build balance query: %v", err)
		tx.Rollback(ctx)
		return err
	}

	err = tx.QueryRow(ctx, sqlStr, args...).Scan(&balance)
	if err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("user not found")
		}
		return err
	}

	if balance < price {
		log.Println("Insufficient funds for user")
		tx.Rollback(ctx)
		return errors.New("insufficient funds")
	}

	updateBalance := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Set("coins", squirrel.Expr("coins - $1", price)).
		Where("id = $2", userId).
		Suffix("RETURNING coins")

	balanceSql, _, err := updateBalance.ToSql()
	if err != nil {
		log.Printf("Failed to build update balance query: %v", err)
		tx.Rollback(ctx)
		return err
	}

	var updatedCoins int64
	err = tx.QueryRow(ctx, balanceSql, price, userId).Scan(&updatedCoins)
	if err != nil {
		log.Printf("Failed to update user's balance: %v", err)
		tx.Rollback(ctx)
		return errors.New("failed to update user's balance")
	}

	inventoryQuery := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("quantity").
		From("inventory").
		Where(squirrel.Eq{"user_id": userId, "item_type": itemType})

	var quantity int64
	inventorySql, inventoryArgs, err := inventoryQuery.ToSql()
	if err != nil {
		log.Printf("Failed to build inventory query: %v", err)
		tx.Rollback(ctx)
		return err
	}

	err = tx.QueryRow(ctx, inventorySql, inventoryArgs...).Scan(&quantity)
	if err != nil {

		insertInventory := squirrel.Insert("inventory").PlaceholderFormat(squirrel.Dollar).
			Columns("user_id", "item_type", "quantity").
			Values(userId, itemType, 1).
			Suffix("RETURNING id")

		inventoryInsertSql, inventoryInsertArgs, err := insertInventory.ToSql()
		if err != nil {
			log.Printf("Failed to build insert inventory query: %v", err)
			tx.Rollback(ctx)
			return err
		}

		var inventoryId int
		err = tx.QueryRow(ctx, inventoryInsertSql, inventoryInsertArgs...).Scan(&inventoryId)
		if err != nil {
			log.Printf("Failed to insert item into inventory: %v", err)
			tx.Rollback(ctx)
			return errors.New("failed to add item to inventory")
		}

	} else {
		updateInventory := squirrel.Update("inventory").
			PlaceholderFormat(squirrel.Dollar).
			Set("quantity", squirrel.Expr("quantity + 1")).
			Where(squirrel.Eq{"user_id": userId, "item_type": itemType}).
			Suffix("RETURNING quantity")

		inventoryUpdateSql, inventoryUpdateArgs, err := updateInventory.ToSql()
		if err != nil {
			log.Printf("Failed to build update inventory query: %v", err)
			tx.Rollback(ctx)
			return err
		}

		err = tx.QueryRow(ctx, inventoryUpdateSql, inventoryUpdateArgs...).Scan(&quantity)
		if err != nil {
			log.Printf("Failed to update inventory: %v", err)
			tx.Rollback(ctx)
			return errors.New("failed to update inventory")
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return err
	}

	return nil
}

func GetUserInfo(userId int) (map[string]interface{}, error) {
	ctx := context.Background()

	var coins int64
	query := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("coins").
		From("users").
		Where(squirrel.Eq{"id": userId})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = Pool.QueryRow(ctx, sqlStr, args...).Scan(&coins)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	inventoryQuery := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("item_type", "quantity").
		From("inventory").
		Where(squirrel.Eq{"user_id": userId})

	sqlStr, args, err = inventoryQuery.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := Pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	inventory := make([]map[string]int64, 0)
	for rows.Next() {
		var itemType string
		var quantity int64
		if err := rows.Scan(&itemType, &quantity); err != nil {
			return nil, err
		}
		inventory = append(inventory, map[string]int64{itemType: quantity})
	}

	transactionsQuery := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("from_user_id", "to_user_id", "amount", "timestamp").
		From("transactions").
		Where(squirrel.Or{
			squirrel.Eq{"from_user_id": userId},
			squirrel.Eq{"to_user_id": userId},
		})

	sqlStr, args, err = transactionsQuery.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err = Pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]map[string]interface{}, 0)
	for rows.Next() {
		var fromUserID, toUserID sql.NullInt64
		var amount int64
		var timestamp time.Time

		if err := rows.Scan(&fromUserID, &toUserID, &amount, &timestamp); err != nil {
			return nil, err
		}

		transaction := map[string]interface{}{
			"from_user_id": fromUserID.Int64,
			"to_user_id":   toUserID.Int64,
			"amount":       amount,
			"timestamp":    timestamp.Format(time.RFC3339),
		}
		transactions = append(transactions, transaction)
	}

	return map[string]interface{}{
		"coins":        coins,
		"inventory":    inventory,
		"transactions": transactions,
	}, nil
}
