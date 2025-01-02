package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"Dana/agent/model"
)

type PrometheusRepo interface {
	AddServerInput(context.Context, *model.Prometheus) error
	GetServers(context.Context) ([]*model.Prometheus, error)
	DeleteServer(context.Context, string) error
}

func NewPrometheusRepo(pg *pgx.Conn) PrometheusRepo {
	_, err := pg.Exec(context.Background(), PrometheusConfigCreateTable)
	if err != nil {
		panic(err)
	}
	return &prometheusRepo{
		pg: pg,
	}
}

type prometheusRepo struct {
	pg *pgx.Conn
}

const (
	PrometheusConfigCreateTable = `
	CREATE TABLE IF NOT EXISTS prometheus_config (
		id SERIAL PRIMARY KEY,
		urls TEXT[] NOT NULL,
		metric_version INT DEFAULT 1,
		timeout INTERVAL DEFAULT '5 seconds'
	);
	`
	PrometheusInsertQuery = `
	INSERT INTO prometheus_config (urls, metric_version, timeout) 
	VALUES ($1, $2, $3)
	RETURNING id;
	`
	PrometheusSelectQuery = `
	SELECT id, urls, metric_version, timeout 
	FROM prometheus_config;
	`
	PrometheusDeleteQuery = `
	DELETE FROM prometheus_config 
	WHERE id = $1;
	`
)

func (p *prometheusRepo) AddServerInput(ctx context.Context, prometheus *model.Prometheus) error {
	var id int
	err := p.pg.QueryRow(ctx, PrometheusInsertQuery, prometheus.URLs, prometheus.MetricVersion, prometheus.Timeout).Scan(&id)
	if err != nil {
		return err
	}
	prometheus.ID = id
	return nil
}

func (p *prometheusRepo) GetServers(ctx context.Context) ([]*model.Prometheus, error) {
	rows, err := p.pg.Query(ctx, PrometheusSelectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*model.Prometheus
	for rows.Next() {
		var prometheus model.Prometheus
		if err := rows.Scan(&prometheus.ID, &prometheus.URLs, &prometheus.MetricVersion, &prometheus.Timeout); err != nil {
			return nil, err
		}
		servers = append(servers, &prometheus)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return servers, nil
}

func (p *prometheusRepo) DeleteServer(ctx context.Context, id string) error {
	commandTag, err := p.pg.Exec(ctx, PrometheusDeleteQuery, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return errors.New("no rows affected")
	}
	return nil
}
