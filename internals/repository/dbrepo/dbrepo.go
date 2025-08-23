package dbrepo

import (
	"github.com/jackc/pgx/v5"
	"github.com/stackninja.pro/goth/internals/repository"
	"github.com/stackninja.pro/goth/src/config"
)

type neonDBRepo struct {
	App *config.AppConfig
	DB  *pgx.Conn
}



func NewPostgresRepo(a *config.AppConfig, conn *pgx.Conn) repository.DatabaseRepo {
	return &neonDBRepo{
		App: a,
		DB:  conn,
	}
}
