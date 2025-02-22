package services

import (
	"context"
	"errors"
	"log"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

func RegisterUser(username, passwordHash string) (int, error) {
	ctx := context.Background()

	insertQuery := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Insert("users").
		Columns("username", "password_hash", "coins").
		Values(username, passwordHash, 1000).
		Suffix("RETURNING id")

	sqlStr, args, err := insertQuery.ToSql()
	if err != nil {
		return 0, err
	}

	var userId int
	err = Pool.QueryRow(ctx, sqlStr, args...).Scan(&userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("user registration failed")
		}
		return 0, err
	}

	return userId, nil
}

func CheckUserExists(username string) (bool, int, string, error) {
	ctx := context.Background()

	query := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("id", "password_hash").
		From("users").
		Where(squirrel.Eq{"username": username})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		log.Printf("Failed to build SQL query: %v", err)
		return false, 0, "", err
	}

	var userId int
	var userHash string
	err = Pool.QueryRow(ctx, sqlStr, args...).Scan(&userId, &userHash)
	if err != nil {
		return false, 0, "", nil
	}

	log.Printf("User found: %s (ID: %d)", username, userId)
	return true, userId, userHash, nil
}
