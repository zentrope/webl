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
	"fmt"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

//-----------------------------------------------------------------------------
// Queries for Public Pages
//-----------------------------------------------------------------------------

type LatestPost struct {
	UUID          string
	DateCreated   time.Time
	DateUpdated   time.Time
	DatePublished time.Time
	Status        string
	Slugline      string
	Text          string
}

func (conn *Database) FocusPost(uuid string) (*LatestPost, error) {
	var query = `
	 select
		 p.uuid, p.date_created, p.date_updated, p.date_published, p.status, p.slugline, p.text
	 from
		 post p
	 where
		 p.uuid = $1
		 and p.status='published'
	`

	rows, err := conn.db.Query(query, uuid)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rows.Next()
	var p LatestPost

	err = rows.Scan(
		&p.UUID,
		&p.DateCreated,
		&p.DateUpdated,
		&p.DatePublished,
		&p.Status,
		&p.Slugline,
		&p.Text,
	)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (conn *Database) LatestPosts(limit int) ([]*LatestPost, error) {
	var query = `
	 select
		 p.uuid, p.date_created, p.date_updated, p.date_published, p.status,
		 p.slugline, p.text
	 from
		 post p
	 where
		 p.status='published'
	 order
		 by p.date_published desc
	 limit $1`

	rows, err := conn.db.Query(query, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := make([]*LatestPost, 0)

	for rows.Next() {
		var p LatestPost
		err := rows.Scan(
			&p.UUID,
			&p.DateCreated,
			&p.DateUpdated,
			&p.DatePublished,
			&p.Status,
			&p.Slugline,
			&p.Text,
		)

		if err != nil {
			return make([]*LatestPost, 0), err
		}

		posts = append(posts, &p)
	}

	return posts, nil
}

type ArchiveEntry struct {
	UUID          string
	DateCreated   time.Time
	DateUpdated   time.Time
	DatePublished time.Time
	Slugline      string
}

func (conn *Database) ArchiveEntries() ([]*ArchiveEntry, error) {

	var query = `
		select
			uuid, date_created, date_updated, date_published, slugline
		from
			post
		where
			status='published'
		order by
			date_published desc;
	`

	rows, err := conn.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	entries := make([]*ArchiveEntry, 0)

	for rows.Next() {
		var e ArchiveEntry
		err := rows.Scan(
			&e.UUID,
			&e.DateCreated,
			&e.DateUpdated,
			&e.DatePublished,
			&e.Slugline,
		)

		if err != nil {
			return make([]*ArchiveEntry, 0), err
		}

		entries = append(entries, &e)
	}

	return entries, nil

}

func (conn *Database) CreatePost(authorUuid, slugline, status, text, datePublished string) (string, error) {
	uuid := mkUuid()
	_, err := conn.db.Exec(
		"insert into post (uuid, author_uuid, slugline, status, text, date_published) values ($1, $2, $3, $4, $5, $6)",
		uuid, authorUuid, slugline, status, text, datePublished)
	return uuid, err
}

//-----------------------------------------------------------------------------
// Queries for GraphQL
//-----------------------------------------------------------------------------

type Post struct {
	UUID          string
	AuthorUuid    string
	DateCreated   time.Time
	DateUpdated   time.Time
	DatePublished time.Time
	Status        string
	Slugline      string
	Text          string
	WordCount     int
}

func (conn *Database) Posts() ([]*Post, error) {
	q := mkPostSql("")
	return conn.postQuery(q)
}

func (conn *Database) PostsByAuthor(authorUuid string) ([]*Post, error) {
	return conn.postQuery(mkPostSql("where author_uuid=$1"), authorUuid)
}

func (conn *Database) Post(uuid string) (*Post, error) {
	posts, err := conn.postQuery(mkPostSql("where uuid=$1"), uuid)
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return nil, fmt.Errorf("Post by UUID[%s] not found.", uuid)
	}

	return posts[0], nil
}

// Delete a post if it also belongs to the indicated author.
func (conn *Database) DeletePost(uuid, authorId string) error {
	q := "delete from post where uuid=$1 and author_uuid=$2"
	_, err := conn.db.Exec(q, uuid, authorId)
	return err
}

type PostStatus int

const (
	PS_Published PostStatus = iota
	PS_Draft
)

// Update a post, assuming the uuid and author match the same post.
func (conn *Database) UpdatePost(uuid, slugline, text, datePublished, authorUuid string) (*Post, error) {

	q := `update post set slugline=$1, text=$2, date_updated=now(), date_published=$3 where uuid=$4
					 and author_uuid = $5`

	_, err := conn.db.Exec(q, slugline, text, datePublished, uuid, authorUuid)
	if err != nil {
		return nil, err
	}

	return conn.Post(uuid)
}

func (conn *Database) SetPostStatus(uuid, authorUuid string, status PostStatus) (*Post, error) {
	s := "draft"
	if status == PS_Published {
		s = "published"
	}

	q := "update post set status=$1, date_updated=now() where uuid=$2 and author_uuid=$3"

	_, err := conn.db.Exec(q, s, uuid, authorUuid)
	if err != nil {
		return nil, err
	}

	return conn.Post(uuid)
}

//-----------------------------------------------------------------------------
// Support
//-----------------------------------------------------------------------------

func mkUuid() string {
	return fmt.Sprintf("%s", uuid.NewV4())
}

func (conn *Database) postQuery(query string, args ...interface{}) ([]*Post, error) {

	rows, err := conn.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	results, err := rowsToPosts(rows)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func mkPostSql(where string) string {
	q := "select uuid, author_uuid, date_created, date_updated, date_published, status, slugline, text from post %s"
	return fmt.Sprintf(q, where)
}

func rowsToPosts(rows *sql.Rows) ([]*Post, error) {

	posts := make([]*Post, 0)

	for rows.Next() {
		post, err := rowToPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func rowToPost(rows *sql.Rows) (*Post, error) {
	var p Post
	err := rows.Scan(
		&p.UUID,
		&p.AuthorUuid,
		&p.DateCreated,
		&p.DateUpdated,
		&p.DatePublished,
		&p.Status,
		&p.Slugline,
		&p.Text,
	)
	p.WordCount = len(strings.Fields(p.Text))
	if err != nil {
		return nil, err
	}
	return &p, nil
}
