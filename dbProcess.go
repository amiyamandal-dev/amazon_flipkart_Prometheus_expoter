package main

import (
	"database/sql"
	"errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func DbInit() {
	if _, err := os.Stat("./foo.db"); errors.Is(err, os.ErrNotExist) {
		db, err := sql.Open("sqlite3", "./foo.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		sqlStmt := `
		CREATE table foo (id INTEGER PRIMARY KEY AUTOINCREMENT,  
						productname VARCHAR(64) NULL,
						price REAL NULL,
						website VARCHAR(64) NULL,
						CheckedIn DATE NULL)
		`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
			return
		}

	}

}

func checkerror(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Insert(Product ProductPrice) {
	db, err := sql.Open("sqlite3", "./foo.db")
	defer db.Close()
	checkerror(err)
	stmt, err := db.Prepare("INSERT INTO foo(productname, price, website, CheckedIn) values(?,?,?,?)")
	checkerror(err)
	res, err := stmt.Exec(Product.ProductName, Product.PrductPrice, Product.Website, Product.CheckedIn.String())
	checkerror(err)
	id, err := res.LastInsertId()
	checkerror(err)
	log.Println("inserted Value :- ", id)
}
