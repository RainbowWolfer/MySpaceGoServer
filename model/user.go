package model

import (
	"database/sql"
	"errors"
	"fmt"
)

type User struct {
	ID                 int
	Username           string
	Password           string
	Email              string
	ProfileDescription string
	IsFollowing        bool
}

func ReadUser(rows *sql.Rows) (User, error) {
	var user User
	if err := rows.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.ProfileDescription,
		&user.IsFollowing,
	); err != nil {
		return User{}, errors.New("User Convert Error" + err.Error())
	}
	return user, nil
}

func GetUserID(db *sql.DB, email string, pasword string) (int, error) {
	sql := fmt.Sprintf("select u_id from users where u_email = '%s' and u_password = '%s';", email, pasword)
	println(sql)
	rows, err := db.Query(sql)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	if !rows.Next() {
		return -1, nil
	}
	var userID int
	err = rows.Scan(&userID)
	if err != nil {
		return -1, err
	}
	return userID, nil
}
