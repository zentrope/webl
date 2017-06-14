// Copyright 2017 Keith Irwin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"encoding/json"
	"io/ioutil"

	"github.com/imdario/mergo"
)

//-----------------------------------------------------------------------------
// Exports
//-----------------------------------------------------------------------------

const (
	SITE_BASEURL     = "webl.baseurl"
	SITE_TITLE       = "webl.title"
	SITE_DESCRIPTION = "webl.description"
	SITE_JWT_SECRET  = "webl.jtw.secret"
)

type StorageConfig struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Database string `json:"database,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     string `json:"port,omitempty"`
}

type WebConfig struct {
	Port   string `json:"port,omitempty"`
	Listen string `json:"listen,omitempty"`
}

type SiteConfig struct {
	BaseURL     string
	Title       string
	Description string
	JwtSecret   string
}

type AppConfig struct {
	Storage StorageConfig `json:"storage,omitempty"`
	Web     WebConfig     `json:"web,omitempty"`
	Site    SiteConfig
}

const DefaultConfigFile = "resources/config.json"

func LoadConfigFile(pathToOverride string, resources *Resources) (*AppConfig, error) {

	config, err := loadDefaultConfigFile(resources)
	if err != nil {
		return nil, err
	}

	if pathToOverride == DefaultConfigFile {
		return &config, nil
	}

	override, err := loadConfigFile(pathToOverride)
	if err != nil {
		return nil, err
	}

	if err := mergo.Merge(&override, config); err != nil {
		return nil, err
	}

	return &override, nil
}

func (conn *Database) GetSiteConfig() (SiteConfig, error) {
	q := "select key, value from config"

	rows, err := conn.db.Query(q)
	if err != nil {
		return SiteConfig{}, err
	}

	defer rows.Close()

	site := make(map[string]string, 0)

	for rows.Next() {
		var key string
		var value string

		err := rows.Scan(&key, &value)
		if err != nil {
			return SiteConfig{}, err
		}

		site[key] = value
	}

	return SiteConfig{
		BaseURL:     site[SITE_BASEURL],
		JwtSecret:   site[SITE_JWT_SECRET],
		Title:       site[SITE_TITLE],
		Description: site[SITE_DESCRIPTION],
	}, nil
}

func (conn *Database) UpdateSite(title, description, url string) (SiteConfig, error) {
	kvs := make(map[string]string, 0)

	kvs["webl.title"] = title
	kvs["webl.description"] = description
	kvs["webl.baseurl"] = url

	tx, err := conn.db.Begin()
	if err != nil {
		return SiteConfig{}, nil
	}

	q := "update config set value=$1 where key=$2"

	for k, v := range kvs {
		_, err := conn.db.Exec(q, v, k)
		if err != nil {
			tx.Rollback()
			return SiteConfig{}, err
		}
	}

	tx.Commit()

	return conn.GetSiteConfig()
}

func (conn *Database) AppendSiteConfig(config *AppConfig) (*AppConfig, error) {
	site, err := conn.GetSiteConfig()
	if err != nil {
		return nil, err
	}

	config.Site = site
	return config, nil
}

//-----------------------------------------------------------------------------
// Implementation
//-----------------------------------------------------------------------------

func loadDefaultConfigFile(resources *Resources) (AppConfig, error) {
	contents, err := resources.Private.String("config.json")

	if err != nil {
		return AppConfig{}, err
	}

	var config AppConfig
	if err := json.Unmarshal([]byte(contents), &config); err != nil {
		return AppConfig{}, err
	}

	return config, nil
}

func loadConfigFile(path string) (AppConfig, error) {

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}

	var config AppConfig
	if err := json.Unmarshal(contents, &config); err != nil {
		return AppConfig{}, err
	}

	return config, nil
}
