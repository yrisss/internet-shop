package main

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		os.Exit(1)
	}
	defer db.Close()

	orders := os.Args[1:]
	if len(orders) == 0 {
		fmt.Println("Usage: go run main.go <order_ids>")
		os.Exit(1)
	}

	shelvesOrdersMap := make(map[string]map[int][]string)

	query := `
		SELECT od.order_id, p.name, p.product_id, s.name AS shelf_name, od.count
		FROM OrderDetails od
		JOIN Products p ON od.product_id = p.product_id
		JOIN Shelves s ON p.main_shelf_id = s.shelf_id
		WHERE od.order_id IN (` + strings.Join(orders, ",") + `)
		ORDER BY od.order_id
	`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Error querying database:", err)
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var orderID, productID, count int
		var productName, shelfName string
		err := rows.Scan(&orderID, &productName, &productID, &shelfName, &count)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		additionalShelves, err := getAdditionalShelves(db, productID)
		if err != nil {
			fmt.Println("Error getting additional shelves:", err)
			continue
		}

		productInfo := fmt.Sprintf("%s (id=%d)\nзаказ %d, %d шт", productName, productID, orderID, count)
		if len(additionalShelves) > 0 {
			productInfo += "\nдоп стеллаж: " + strings.Join(additionalShelves, ",")
		}

		if _, ok := shelvesOrdersMap[shelfName]; !ok {
			shelvesOrdersMap[shelfName] = make(map[int][]string)
		}
		shelvesOrdersMap[shelfName][orderID] = append(shelvesOrdersMap[shelfName][orderID], productInfo)
	}

	fmt.Println("=+=+=+=")
	fmt.Printf("Страница сборки заказов %s\n\n", strings.Join(orders, ","))

	var sortedShelves []string
	for shelfName := range shelvesOrdersMap {
		sortedShelves = append(sortedShelves, shelfName)
	}
	sort.Strings(sortedShelves)

	for _, shelfName := range sortedShelves {
		fmt.Printf("===Стеллаж %s\n", shelfName)

		var orderIDs []int
		for orderID := range shelvesOrdersMap[shelfName] {
			orderIDs = append(orderIDs, orderID)
		}
		sort.Ints(orderIDs)

		for _, orderID := range orderIDs {
			for _, product := range shelvesOrdersMap[shelfName][orderID] {
				fmt.Println(product)
				fmt.Println()
			}
		}

		fmt.Println()
	}
}

func getAdditionalShelves(db *sql.DB, productID int) ([]string, error) {
	query := `
		SELECT s.name
		FROM Shelves s
		JOIN ProductSecondaryShelves ps ON s.shelf_id = ps.shelf_id
		WHERE ps.product_id = ?
	`

	rows, err := db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var additionalShelves []string
	for rows.Next() {
		var shelfName string
		err := rows.Scan(&shelfName)
		if err != nil {
			return nil, err
		}
		additionalShelves = append(additionalShelves, shelfName)
	}

	return additionalShelves, nil
}
