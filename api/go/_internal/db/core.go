package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/trinodb/trino-go-client/trino"

	"github/huangc28/kikichoice-be/api/go/_internal/configs"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"go.uber.org/fx"
)

type Conn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Preparex(query string) (*sqlx.Stmt, error)
	Rebind(query string) string
}

func getPostgresqlDSN(cfg *configs.Config) string {
	pgdsn := "postgres://%s:%s@%s:%s/%s"
	params := ""

	if cfg.ENV == configs.Production {
		params = "?sslmode=require&pool_mode=transaction"
	}

	return fmt.Sprintf(pgdsn,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	) + params
}

func NewSQLXPostgresDB(lc fx.Lifecycle, config *configs.Config) *sqlx.DB {
	dsn := getPostgresqlDSN(config)

	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		panic(fmt.Errorf("failed to parse pgx config: %w", err))
	}

	pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	sqlDB := stdlib.OpenDB(*pgxConfig)

	driver := sqlx.NewDb(sqlDB, "pgx")
	driver.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error { return driver.Close() },
	})

	return driver
}
