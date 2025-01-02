package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"Dana/agent/model"
)

type SnmpRepo interface {
	AddSnmpInput(context.Context, *model.Snmp) error
	GetSnmps(context.Context) ([]*model.Snmp, error)
	DeleteSnmp(context.Context, string) error
}

func NewSnmpRepo(pg *pgx.Conn) SnmpRepo {
	_, err := pg.Exec(context.Background(), SnmpConfigCreateTable)
	if err != nil {
		panic(err)
	}
	return &snmpRepo{
		pg: pg,
	}
}

type snmpRepo struct {
	pg *pgx.Conn
}

const (
	SnmpConfigCreateTable = `
	CREATE TABLE IF NOT EXISTS snmp_config (
		id SERIAL PRIMARY KEY,
		service_address TEXT NOT NULL,
		timeout TEXT NOT NULL,
		version TEXT NOT NULL,
		security_info JSONB NOT NULL
	);
	`
	SnmpInsertQuery = `
	INSERT INTO snmp_config (service_address, timeout, version, security_info)
	VALUES ($1, $2, $3, $4)
	RETURNING id;
	`
	SnmpSelectQuery = `
	SELECT id, service_address, timeout, version, security_info
	FROM snmp_config;
	`
	SnmpDeleteQuery = `
	DELETE FROM snmp_config 
	WHERE id = $1;
	`
)

func (s *snmpRepo) AddSnmpInput(ctx context.Context, snmp *model.Snmp) error {
	// Security-related information as JSONB
	securityInfo := map[string]string{
		"sec_name":      snmp.SecName,
		"auth_protocol": snmp.AuthProtocol,
		"auth_password": snmp.AuthPassword,
		"sec_level":     snmp.SecLevel,
		"priv_protocol": snmp.PrivProtocol,
		"priv_password": snmp.PrivPassword,
	}

	var id int
	err := s.pg.QueryRow(ctx, SnmpInsertQuery, snmp.ServiceAddress, snmp.Timeout, snmp.Version, securityInfo).Scan(&id)
	if err != nil {
		return err
	}
	snmp.ID = id
	return nil
}

func (s *snmpRepo) GetSnmps(ctx context.Context) ([]*model.Snmp, error) {
	rows, err := s.pg.Query(ctx, SnmpSelectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snmps []*model.Snmp
	for rows.Next() {
		var snmp model.Snmp
		var securityInfo map[string]string
		if err := rows.Scan(&snmp.ID, &snmp.ServiceAddress, &snmp.Timeout, &snmp.Version, &securityInfo); err != nil {
			return nil, err
		}
		// Mapping the security info back to struct fields
		snmp.SecName = securityInfo["sec_name"]
		snmp.AuthProtocol = securityInfo["auth_protocol"]
		snmp.AuthPassword = securityInfo["auth_password"]
		snmp.SecLevel = securityInfo["sec_level"]
		snmp.PrivProtocol = securityInfo["priv_protocol"]
		snmp.PrivPassword = securityInfo["priv_password"]
		snmps = append(snmps, &snmp)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return snmps, nil
}

func (s *snmpRepo) DeleteSnmp(ctx context.Context, id string) error {
	commandTag, err := s.pg.Exec(ctx, SnmpDeleteQuery, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return errors.New("no rows affected")
	}
	return nil
}
