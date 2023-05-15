package pgconn

import (
	"context"
	"fmt"
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

// Удаляет базу данных
func DeleteDatabase(ctx context.Context, conn *pgx.Conn, db Database) error {
	_, err := conn.Exec(ctx, "DROP DATABASE $1", db.Name) // ошибка с доларом
	if err != nil {
		return err
	}
	return nil
}
