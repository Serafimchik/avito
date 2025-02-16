package services

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx"
)

func SendCoins(fromUserId, toUserId int, amount int64) error {
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

	var balance int64
	err = tx.QueryRow(ctx, "SELECT coins FROM users WHERE id = $1", fromUserId).Scan(&balance)
	if err != nil {
		log.Printf("Failed to scan sender's balance: %v", err)
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("sender not found")
		}
		return err
	}
	if balance < amount {
		log.Println("Insufficient funds for sender")
		tx.Rollback(ctx)
		return errors.New("insufficient funds")
	}

	var updatedCoins int64
	err = tx.QueryRow(ctx, "UPDATE users SET coins = coins - $1 WHERE id = $2 RETURNING coins", amount, fromUserId).Scan(&updatedCoins)
	if err != nil {
		log.Printf("Failed to update sender's balance: %v", err)
		tx.Rollback(ctx)
		return errors.New("failed to update sender's balance")
	}
	log.Printf("Sender's balance updated successfully. New balance: %d", updatedCoins)

	err = tx.QueryRow(ctx, "UPDATE users SET coins = coins + $1 WHERE id = $2 RETURNING coins", amount, toUserId).Scan(&updatedCoins)
	if err != nil {
		log.Printf("Failed to update recipient's balance: %v", err)
		tx.Rollback(ctx)
		return errors.New("failed to update recipient's balance")
	}
	log.Printf("Recipient's balance updated successfully. New balance: %d", updatedCoins)

	var transactionId int
	err = tx.QueryRow(ctx, "INSERT INTO transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3) RETURNING id", fromUserId, toUserId, amount).Scan(&transactionId)
	if err != nil {
		log.Printf("Failed to create transaction record: %v", err)
		tx.Rollback(ctx)
		return errors.New("failed to create transaction record")
	}
	log.Printf("Transaction created successfully. Transaction ID: %d", transactionId)

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return err
	}
	return nil
}
