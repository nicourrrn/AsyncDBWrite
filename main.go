package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func GetRowId(db *sql.DB, selectQuery, insertQuery string, args ...interface{}) int64 {
	row := db.QueryRow(selectQuery, args...)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println(err)
		}
		result, err := db.Exec(insertQuery, args...)
		if err != nil {
			if strings.HasPrefix(err.Error(), "Error 1062") {
				return GetRowId(db, selectQuery, insertQuery, args...)
			}
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
	db, err := sql.Open("mysql", "student:1234@/fullstack_shop")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	wg.Add(pool.Count)
	for i := 0; i < pool.Count; i++ {
		fmt.Println("start")
		go pool.Run(&wg, func(shop Shop) {
			shopTypeId := GetRowId(db, "SELECT id FROM shop_types WHERE name = ?",
				"INSERT INTO shop_types(name) VALUE (?)", shop.Type)
			_, err = db.Exec("INSERT INTO shops VALUE (?, ?, ?, ?, ?, ?)",
				shop.Id, shop.Name, shopTypeId, shop.Image, shop.WorkingHours.Opening,
				shop.WorkingHours.Closing)

			if err != nil {
				log.Println(err)
			}

			for _, prod := range shop.Menu {
				prodTypeId := GetRowId(db, "SELECT id FROM prod_types WHERE name = ?",
					"INSERT INTO prod_types(name) VALUE (?)", prod.Type)
				_, err = db.Exec(
					"INSERT INTO products VALUE (?, ?, ?, ?, ?, ?)",
					prod.Id, prod.Name, prod.Price, prod.Image, prodTypeId, shop.Id)

				if err != nil {
					log.Println(err)
				}

				for _, ing := range prod.Ingredients {
					ingId := GetRowId(db, "SELECT id FROM ingredients WHERE name = ?",
						"INSERT INTO ingredients(name) VALUE (?)", ing)
					_, err = db.Exec("INSERT INTO prod_ingredient VALUE (?, ?)", prod.Id, ingId)

					if err != nil {
						log.Println(err)
					}
				}
			}
		})
	}

	http.HandleFunc("/upload", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "not supported", http.StatusBadRequest)
			return
		}
		var shop Shop
		err := json.NewDecoder(request.Body).Decode(&shop)
		if err != nil {
			http.Error(writer, "not parsed", http.StatusBadRequest)
			return
		}
		request.Body.Close()
		pool.Sender <- shop
		writer.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/exit", func(writer http.ResponseWriter, request *http.Request) {
		pool.Stop()
		wg.Wait()
		writer.WriteHeader(http.StatusContinue)
		os.Exit(1)
	})
	//shops := make([]Shop, 0)
	//var shop Shop
	//dir, err := os.Open("shops")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//files, err := dir.Readdir(-1)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//for _, i := range files {
	//	shop, err = NewFromJson("shops/" + i.Name())
	//	if err != nil {
	//		log.Println(err)
	//	}
	//	shops = append(shops, shop)
	//}
	//
	//for _, shop = range shops {
	//	pool.Sender <- shop
	//}
	log.Println(http.ListenAndServe("localhost:8080", nil))
	pool.Stop()
	wg.Wait()
}
