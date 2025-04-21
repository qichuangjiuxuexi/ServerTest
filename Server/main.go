package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const saveFile = "saves/player_data.json"

func savePlayerData(filename string, data interface{}) error {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, file, 0644)
	return err
}

func loadPlayerData(filename string, data interface{}) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, data)
	if err != nil {
		return err
	}
	return nil
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// TODO: Get data from request body
	data := map[string]interface{}{"message": "This is a test save"}

	err := savePlayerData(filename, data)
	if err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Data saved to %s successfully!\n", filename)
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	var data map[string]interface{}
	err := loadPlayerData(filename, &data)
	if err != nil {
		http.Error(w, "Failed to load data", http.StatusInternalServerError)
		return
	}

	// TODO: Send data to response body
	fmt.Fprintf(w, "Data loaded from %s successfully!\n", filename)
	fmt.Fprintf(w, "Data: %+v\n", data)
}

func main() {
	http.HandleFunc("/save", saveHandler)
	http.HandleFunc("/load", loadHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server is running. Use /save?file=filename to save data and /load?file=filename to load data.")
	})

	http.HandleFunc("/testsave", testSaveHandler)

	fmt.Println("Server listening on port 8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func testSaveHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	defaultData := `{"message": "This is a default value"}`
	filePath := "saves/" + filename
	// Ensure the "saves" directory exists
	if _, err := os.Stat("saves"); os.IsNotExist(err) {
		os.Mkdir("saves", 0755) // Create the directory if it doesn't exist
	}
	err := ioutil.WriteFile(filePath, []byte(defaultData), 0644)
	if err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Data saved to %s successfully!\n", filename)
}
