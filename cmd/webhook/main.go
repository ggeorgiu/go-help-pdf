package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("server started")
	http.HandleFunc("/listen", handle)

	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		fmt.Printf("failed to listen, %v", err)
	}
}

func handle(writer http.ResponseWriter, request *http.Request) {
	fmt.Println(request)
	writer.WriteHeader(http.StatusOK)
}
