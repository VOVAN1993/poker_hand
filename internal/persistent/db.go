package persistent

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	db struct {
		pool      *pgxpool.Pool
		closeFunc func()
		mutex     *sync.Mutex
	}
	Persistent interface {
		Start(context.Context) error
		Stop()
	}
)

func NewPersistent() Persistent {
	return &db{mutex: &sync.Mutex{}}
}

func (db *db) Start(ctx context.Context) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	if db.pool != nil {
		return fmt.Errorf("db already initialized")
	}
	pswd := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	urlExample := fmt.Sprintf("postgres://postgres:%s@host.docker.internal:5433/%s", pswd, dbName)
	dbpool, err := pgxpool.New(ctx, os.Getenv(urlExample))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		return err
	}
	db.pool = dbpool
	db.closeFunc = dbpool.Close
	return nil
}

func (db *db) Stop() {
	db.closeFunc()
}

func (db *db) Tst(ctx context.Context) error {
	pswd := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	urlExample := fmt.Sprintf("postgres://postgres:%s@host.docker.internal:5433/%s", pswd, dbName)
	dbpool, err := pgxpool.New(ctx, os.Getenv(urlExample))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		return err
	}
	defer dbpool.Close()

	conn, err := pgx.Connect(ctx, urlExample)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select id from users")
	if err != nil {
		return err
	}

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		fmt.Println(name)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}
