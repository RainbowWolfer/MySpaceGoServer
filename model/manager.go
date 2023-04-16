package model

import (
	"database/sql"
	"errors"
)

type Manager struct {
	ID       string
	Username string
	Password string
}

func ReadManager(rows *sql.Rows) (Manager, error) {
	var manager Manager
	if err := rows.Scan(
		&manager.ID,
		&manager.Username,
		&manager.Password,
	); err != nil {
		return Manager{}, errors.New("Manager Convert Error " + err.Error())
	}
	return manager, nil
}
