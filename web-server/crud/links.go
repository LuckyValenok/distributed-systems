package crud

import (
	"database/sql"
	"errors"
	"log"
)

var (
	ErrNotFoundObject = errors.New("object not found")
)

func GetLinks(db *sql.DB) ([]string, error) {
	var res string
	var links []string

	rows, err := db.Query("SELECT value FROM links")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&res); err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		links = append(links, res)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func GetLink(db *sql.DB, id int) (string, error) {
	var res string

	query, err := db.Query(`SELECT value FROM links WHERE id = $1`, id)
	if err != nil {
		return "", err
	}
	defer query.Close()

	if !query.Next() {
		return "", ErrNotFoundObject
	}

	if err := query.Scan(&res); err != nil {
		return "", err
	}

	return res, nil
}

func AddLink(db *sql.DB, link string) (int, error) {
	var id int
	err := db.QueryRow(`INSERT INTO links(value) VALUES($1) RETURNING id`, link).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func UpdateLinkStatus(db *sql.DB, id, status int) error {
	stmt, err := db.Prepare(`UPDATE links SET status = $1 WHERE id = $2`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, id)
	if err != nil {
		return err
	}

	return nil
}
