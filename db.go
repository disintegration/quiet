package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/lib/pq"
)

var Db *sql.DB

func DbConnect() {
	var err error

	DATABASE_URL := os.Getenv("DATABASE_URL")
	if DATABASE_URL == "" {
		DATABASE_URL = "postgres://totoro:catbus@localhost:5432/quiet?sslmode=disable"
	}

	connStr, err := pq.ParseURL(DATABASE_URL)
	if err != nil {
		log.Fatal(err)
	}

	Db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

func DbInitSchema() {
	_, err := Db.Exec(`

		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			persona CHARACTER VARYING(100) UNIQUE,
			username CHARACTER VARYING(100) UNIQUE,
			realname CHARACTER VARYING(100) DEFAULT '',
			tm TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS contacts (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id),
			contact_id BIGINT REFERENCES users(id),
			tm TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT unique_user_contact UNIQUE (user_id, contact_id)
		);

		CREATE TABLE IF NOT EXISTS photos (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id),
			rand_id CHARACTER VARYING(20) NOT NULL,
			tm TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			processed NUMERIC(2) DEFAULT 0,
			title CHARACTER VARYING(100) DEFAULT '',
			description TEXT DEFAULT '',
			views_count INTEGER DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS favorites (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id),
			photo_id BIGINT REFERENCES photos(id),
			tm TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT unique_user_photo UNIQUE (user_id, photo_id)
		);

		CREATE TABLE IF NOT EXISTS comments (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES users(id),
			photo_id BIGINT REFERENCES photos(id),
			comment TEXT NOT NULL,
			tm TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func DbDropSchema() {
	_, err := Db.Exec(`   
		DROP TABLE IF EXISTS users CASCADE;
		DROP TABLE IF EXISTS contacts CASCADE;
		DROP TABLE IF EXISTS photos CASCADE;
		DROP TABLE IF EXISTS favorites CASCADE;
		DROP TABLE IF EXISTS comments CASCADE;
	`)
	if err != nil {
		log.Fatal(err)
	}
}
