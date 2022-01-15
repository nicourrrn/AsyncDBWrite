package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"sync"
)

func GetRowId(db *sql.DB, selectQuery, insertQuery string, args ...interface{}) int64 {
	row := db.QueryRow(selectQuery, args...)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		log.Println(err)
		result, err := db.Exec(insertQuery, args)
		if err != nil {
			log.Fatalln(err)
		}
		id, err = result.LastInsertId()
		if err != nil {
			log.Fatalln(err)
		}
	}
	return id
}

func main() {
	pool := NewWorkerPool(4)
	wg := sync.WaitGroup{}
	db, err := sql.Open("mysql", "sorokin:1234@/test")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	wg.Add(pool.Count)
	for i := 0; i < pool.Count; i++ {
		fmt.Println("start")
		go pool.Run(&wg, func(shop Shop) {

			row := db.QueryRow("SELECT id FROM shop_types WHERE name = ?", shop.Type)
			var shopTypeId int64
			err := row.Scan(&shopTypeId)
			if err != nil {
				log.Println(err)
				result, err := db.Exec("INSERT INTO shop_types(name) VALUE (?)", shop.Type)
				if err != nil {
					log.Println(err)
					return
				}
				shopTypeId, err = result.LastInsertId()
				if err != nil {
					log.Println(err)
					return
				}
			}
			_, err = db.Exec("INSERT INTO shops VALUE (?, ?, ?, ?, ?, ?)",
				shop.Id, shop.Name, shopTypeId, shop.Image, shop.WorkingHours.Opening,
				shop.WorkingHours.Closing)
			if err != nil {
				log.Println(err)
				return
			}
			for _, prod := range shop.Menu {
				row := db.QueryRow("SELECT id FROM prod_types WHERE name = ?", prod.Type)
				fmt.Println(prod.Type)
				var prodTypeId int64
				err := row.Scan(&prodTypeId)
				if err != nil {
					log.Println("Product type: ", err)
					result, err := db.Exec(
						"INSERT INTO prod_types(name) VALUE (?)",
						prod.Type)
					if err != nil {
						log.Println(err)
						return
					}
					prodTypeId, err = result.LastInsertId()
					if err != nil {
						log.Println(err)
						return
					}
				}

				_, err = db.Exec(
					"INSERT INTO products VALUE (?, ?, ?, ?, ?, ?)",
					prod.Id, prod.Name, prod.Price, prod.Image, prodTypeId, shop.Id)
				for _, ing := range prod.Ingredients {
					row := db.QueryRow("SELECT id FROM ingredients WHERE name = ?", ing)
					var ingId int64
					err := row.Scan(prodTypeId)
					if err != nil {
						result, err := db.Exec("INSERT INTO ingredients(name) VALUE (?)", ing)
						if err != nil {
							log.Println(err)
							return
						}
						ingId, err = result.LastInsertId()
						if err != nil {
							log.Println(err)
							return
						}
					}
					_, err = db.Exec("INSERT INTO prod_ingredient VALUE (?, ?)", prod.Id, ingId)
				}
			}
		})
	}
	shop, err := NewFromJson("1.json")
	if err != nil {
		fmt.Println(err)
	}
	pool.Sender <- shop
	pool.Stop()
	wg.Wait()
}
