package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// OrderReq represents the format of an incomming order.
type orderReq struct {
	TableNo int64
	Items   []struct {
		ItemID   int64
		Quantity int64
	}
}

type item struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func insertOrderReq(db *sql.DB, order *orderReq) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	stmt, err := tx.Prepare("insert into orders(item_id, quantity, table_no) values (?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, item := range order.Items {
		_, err = stmt.Exec(item.ItemID, item.Quantity, order.TableNo)
		if err != nil {
			return
		}
	}

	return
}

// Order route
func Order(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var req orderReq
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = insertOrderReq(db, &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	NotifyListeners()
}

func getMenuRes(db *sql.DB) (items []item, err error) {
	rows, err := db.Query("select id, name, price from menu")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var i item
		err = rows.Scan(&i.ID, &i.Name, &i.Price)
		if err != nil {
			return nil, err
		}

		items = append(items, i)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return
}

// Menu route
func Menu(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	items, err := getMenuRes(db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	log.Printf("items = %v", items)
	body, err := json.Marshal(items)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(body)
}

// Curry is a functional curry.
func Curry(db *sql.DB, f func(*sql.DB, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(db, w, r)
	}
}
