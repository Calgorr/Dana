package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"Dana/agent/model"
)

type WmiRepo interface {
	AddWmiInput(context.Context, *model.Wmi) error
	GetWmis(context.Context) ([]*model.Wmi, error)
	DeleteWmi(context.Context, string) error
}

func NewWmiRepo(pg *pgx.Conn) WmiRepo {
	_, err := pg.Exec(context.Background(), WmiConfigCreateTable)
	if err != nil {
		panic(err)
	}
	return &wmiRepo{
		pg: pg,
	}
}

type wmiRepo struct {
	pg *pgx.Conn
}

const (
	WmiConfigCreateTable = `
	CREATE TABLE IF NOT EXISTS wmi_config (
		id SERIAL PRIMARY KEY,
		host TEXT NOT NULL,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		queries JSONB NOT NULL,
		methods JSONB NOT NULL
	);
	`
	WmiInsertQuery = `
	INSERT INTO wmi_config (host, username, password, queries, methods)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id;
	`
	WmiSelectQuery = `
	SELECT id, host, username, password, queries, methods
	FROM wmi_config;
	`
	WmiDeleteQuery = `
	DELETE FROM wmi_config 
	WHERE id = $1;
	`
)

func (w *wmiRepo) AddWmiInput(ctx context.Context, wmi *model.Wmi) error {
	// Queries and Methods as JSONB
	queriesJSON := wmi.Queries
	methodsJSON := wmi.Methods

	var id int
	err := w.pg.QueryRow(ctx, WmiInsertQuery, wmi.Host, wmi.Username, wmi.Password, queriesJSON, methodsJSON).Scan(&id)
	if err != nil {
		return err
	}
	wmi.ID = id
	return nil
}

func (w *wmiRepo) GetWmis(ctx context.Context) ([]*model.Wmi, error) {
	rows, err := w.pg.Query(ctx, WmiSelectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wmis []*model.Wmi
	for rows.Next() {
		var wmi model.Wmi
		var queriesJSON, methodsJSON []byte
		if err := rows.Scan(&wmi.ID, &wmi.Host, &wmi.Username, &wmi.Password, &queriesJSON, &methodsJSON); err != nil {
			return nil, err
		}

		// Assuming you have a function to unmarshal JSON into the appropriate Go structures
		if err := json.Unmarshal(queriesJSON, &wmi.Queries); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(methodsJSON, &wmi.Methods); err != nil {
			return nil, err
		}

		wmis = append(wmis, &wmi)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return wmis, nil
}

func (w *wmiRepo) DeleteWmi(ctx context.Context, id string) error {
	commandTag, err := w.pg.Exec(ctx, WmiDeleteQuery, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return errors.New("no rows affected")
	}
	return nil
}
