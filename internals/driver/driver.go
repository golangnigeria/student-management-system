package driver

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// DB holds the database connection pool
type DB struct {
	Conn *pgx.Conn
}

var dbConn = &DB{}

// ConnectToDB establishes a connection to the database (postgres)
func ConnectToDB(dsn string) (*DB, error) {
	conn, err := NewDatabase(dsn)
	if err != nil {
		return nil, err
	}

	dbConn.Conn = conn

	err = testDBConnection(conn)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

// testDBConnection tests the database connection
func testDBConnection(d *pgx.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return d.Ping(ctx)

}

// NewDatabase creates a new database connection
func NewDatabase(dsn string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	return conn, nil
}
