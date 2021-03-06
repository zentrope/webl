//
// Copyright (c) 2017 Keith Irwin
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package server

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	// The Postgres driver requires an unnamed import
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Author struct {
	Uuid   string
	Name   string
	Email  string
	Type   string
	Status string
}

func (conn *Database) Authentic(email, password string) (*Author, error) {
	const query = "select uuid, password from author where lower(email)=lower($1)"
	rows, err := conn.db.Query(query, email)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("User not found.")
	}

	var hash string
	var authorUuid string
	err = rows.Scan(&authorUuid, &hash)
	if err != nil {
		return nil, err
	}

	decoded, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(decoded, []byte(password))
	if err != nil {
		return nil, err
	}

	return conn.Author(authorUuid)
}

func (conn *Database) UpdateAuthor(authorUuid, name, email string) (*Author, error) {
	q := "update author set name=$1, email=$2 where uuid=$3"
	_, err := conn.db.Exec(q, name, email, authorUuid)
	if err != nil {
		return nil, err
	}
	return conn.Author(authorUuid)
}

func (conn *Database) UpdateAuthorPassword(authorUuid, password string) (*Author, error) {
	raw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	encoded := fmt.Sprintf("%x", raw)

	_, err = conn.db.Exec("update author set password=$1 where uuid=$2", encoded, authorUuid)
	if err != nil {
		return nil, err
	}

	return conn.Author(authorUuid)
}

func (conn *Database) Author(authorUuid string) (*Author, error) {
	const query = "select uuid, name, email, type, status from author where uuid=$1"
	rows, err := conn.db.Query(query, authorUuid)

	defer rows.Close()

	if err != nil {
		return nil, err
	}

	rows.Next()
	return rowToAuthor(rows)
}

func (conn *Database) Authors() ([]*Author, error) {
	rows, err := conn.db.Query("select uuid, name, email, type, status from author")
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	authors := make([]*Author, 0)

	for rows.Next() {
		author, err := rowToAuthor(rows)
		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}

	return authors, nil
}

func rowToAuthor(rows *sql.Rows) (*Author, error) {
	var a Author
	err := rows.Scan(&a.Uuid, &a.Name, &a.Email, &a.Status, &a.Type)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
