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

		CreateTournamentsTable(ctx context.Context) error

		SaveTournaments(ctx context.Context, t Tournament) (bool, error)
		ListTournaments(ctx context.Context, whereOpts ...WhereOpt) ([]Tournament, error)
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
	dbAddr := os.Getenv("DB_ADDR")
	urlExample := fmt.Sprintf("postgres://postgres:%s@%s:5433/%s", pswd, dbAddr, dbName)
	dbpool, err := pgxpool.New(ctx, urlExample)
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

func (db *db) ListTournaments(ctx context.Context, whereOpts ...WhereOpt) ([]Tournament, error) {
	query := `
		SELECT * FROM tournaments
	`
	where := constructsOption(whereOpts...)
	if where.ID != nil {
		query += fmt.Sprintf(" WHERE id = '%s'", *where.ID)
	}
	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var tournamets []Tournament
	for rows.Next() {
		var t Tournament
		if err := rows.Scan(&t.ID, &t.BI, &t.Players, &t.TotalPrizePool,
			&t.Started, &t.MyPlace, &t.MyPrize, &t.Reentries, &t.Name, &t.Type); err != nil {
			return nil, err
		}
		tournamets = append(tournamets, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tournamets, nil
}

func (db *db) SaveTournaments(ctx context.Context, t Tournament) (bool, error) {

	query := `
	INSERT INTO tournaments (
		id, bi, players, total_prize_pool, started, my_place, my_prize, reentries, name, type
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
	)ON CONFLICT (id) DO NOTHING;`

	st, err := db.pool.Exec(ctx, query,
		t.ID,
		t.BI,
		t.Players,
		t.TotalPrizePool,
		t.Started,
		t.MyPlace,
		t.MyPrize,
		t.Reentries,
		t.Name,
		t.Type,
	)
	if err != nil {
		return false, fmt.Errorf("failed to insert tournament: %w", err)
	}

	return st.RowsAffected() > 0, nil
}

func (db *db) CreateTournamentsTable(ctx context.Context) error {
	q := `CREATE TABLE IF NOT EXISTS tournaments (
		id TEXT PRIMARY KEY,
		bi FLOAT4 NOT NULL,
		players INT NOT NULL,
		total_prize_pool FLOAT4 NOT NULL,
		started TIMESTAMP NOT NULL,
		my_place INT NOT NULL,
		my_prize FLOAT4 NOT NULL,
		reentries INT NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL
	);`

	_, err := db.pool.Exec(ctx, q)
	if err != nil {
		return fmt.Errorf("failed to create tournaments table: %w", err)
	}

	return nil
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
