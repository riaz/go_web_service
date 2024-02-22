// main_test.go
package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	main "github.com/riaz/go_web_service"
)

var a main.App

func TestMain(m *testing.M) {
	fmt.Println("Hello")
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func TestEmptyTables(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistantProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	// we will check the response for matching the ok status
	checkResponseCode(t, http.StatusNotFound, response.Code) // since the product doesnt exist

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["error"] != "Product not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'", m["error"])
	}

}

func TestCreateProduct(t *testing.T) {

	clearTable()

	var jsonStr = []byte(`{"name": "test product", "price": 11.22}`)

	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	// getting a map from string to interface
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got %v", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}

	// the id is compared to 1.0 because JSON marshalling converts numbers to floats
	// when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}
}

// Getting a Product
func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

// Test update product
func TestUpdateProduct(t *testing.T) {
	// We will fetch a entry, modify it and persist it.
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// next we will unmarshall the response
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	// preparing updated fields
	var jsonStr = []byte(`{"name": "test product - updated name", "price": 11.22}`) // note: the price is still the same
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}

	// unmarshalling the response
	json.Unmarshal(response.Body.Bytes(), &m)

	// Doing a get again
	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// getting the new updated values
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	// we will do if the update happened as expected
	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected  %v got %v", originalProduct["id"], m["id"])
	}

	if m["name"] != originalProduct["name"] {
		t.Errorf("Expected %v got  %v", originalProduct["name"], m["name"])
	}

	if m["price"] != originalProduct["price"] {
		t.Errorf("Expected %v got %v", originalProduct["price"], m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	// we added a product and new we want to delete it.
	req, _ := http.NewRequest("DELETE", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	// we can verify, that it should not exist anymore
	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO products(name,price) VALUES ($1, $2)", "Product"+strconv.Itoa(i), (i+1.0)*10)
	}
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`

func executeRequest(req *http.Request) *httptest.ResponseRecorder {

	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d got %d", expected, actual)
	}
}
