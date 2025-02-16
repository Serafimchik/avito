package services

import (
	"context"
	"log"

	"github.com/Masterminds/squirrel"
)

func RegisterUser(username, passwordHash string) error {
	ctx := context.Background()

	var userId int

	insertQuery := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Insert("users").
		Columns("username", "password_hash", "coins").
		Values(username, passwordHash, 1000).
		Suffix("RETURNING id")

	sqlStr, args, err := insertQuery.ToSql()
	if err != nil {
		return err
	}

	err = Pool.QueryRow(ctx, sqlStr, args...).Scan(&userId)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		return err
	}

	return nil
}
