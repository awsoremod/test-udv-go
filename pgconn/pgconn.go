package pgconn

import (
	"context"
	"fmt"
	"strings"
	"test-udv/pgpass"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	Name string `db:"datname"`
}

// Открывает соединение к базе данных postgres
func OpenConnect(ctx context.Context, config *pgpass.Entry) (*pgx.Conn, error) {
	connString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s target_session_attrs=read-write",
		config.Host, config.Port, config.Dbname, config.User, config.Password)

	connConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Получает версию базы данных postgres
func GetVersion(ctx context.Context, conn *pgx.Conn) error {
	var version string
	err := conn.QueryRow(ctx, "select version()").Scan(&version)
	if err != nil {
		return err
	}
	return nil
}

// Получает список баз данных
func DatabaseList(ctx context.Context, conn *pgx.Conn) ([]Database, error) {
	query := `SELECT datname FROM pg_database`
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	databases, err := pgx.CollectRows(rows, pgx.RowToStructByName[Database])
	if err != nil {
		return nil, err
	}

	return databases, nil
}

// Удаляет базу данных.
// Нельзя удалить базу данных:
// 1) если есть активное подключение к базе данных;
// 2) если вы не являетесь владельцем базы данных.
func DeleteDatabase(ctx context.Context, conn *pgx.Conn, db Database) error {
	dbName := quoteIdentifier(db.Name)
	_, err := conn.Exec(ctx, "DROP DATABASE "+dbName)
	if err != nil {
		return err
	}
	return nil
}

func quoteIdentifier(s string) string {
	return `"` + strings.Replace(s, `"`, `""`, -1) + `"`
}
