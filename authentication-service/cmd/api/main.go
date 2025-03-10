package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"
var counts int64

type Config struct {
	Repo data.Repository
	Client *http.Client
}

func main() {
	log.Println("starting auth service")

	// connect db
	conn := connectToDB()
	if conn == nil {
		log.Panic("could not connect to db")
	}

	// set up config
	app := Config{
		Client: &http.Client{},
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

func openDB(dsn string) (*sql.DB, error){
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("postgres not ready yet...")
			counts++
		} else {
			log.Println("connected to postgres!!")
			return connection
		}
		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("backing off for 2 secs")
		time.Sleep(2 * time.Second)
		continue
	}
}

func (app *Config) setupRepo(conn *sql.DB) {
	db := data.NewPostgresRepository(conn)
	app.Repo = db
}