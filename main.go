package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	log "github.com/sirupsen/logrus"
)

type Todo struct {
	title string
	date  string
	idNum int
}

var idNums = 1

var todoMap = make(map[int]Todo)

// Helper date function
func getNowDateTime() string {
	return time.Now().Format("Mon Jan 2 15:04:05")
}

// Wrap handlers in a function so I can pass an argument to them
func hello(statsd *statsd.Client) http.HandlerFunc {
	log.Info("hello log Liseth")
	statsd.Incr("liseth_hello_request", []string{"environment:dev"}, 1)
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello\n")
	}
}

func getHeaders(statsd *statsd.Client) http.HandlerFunc {
	statsd.Incr("liseth_get_headers_request", []string{"environment:dev"}, 1)
	log.Info("Get headers call")
	return func(w http.ResponseWriter, req *http.Request) {
		for name, headers := range req.Header {
			for _, h := range headers {
				fmt.Fprintf(w, "%v: %v\n", name, h)
			}
		}
	}
}

func getTodos(statsd *statsd.Client) http.HandlerFunc {
	statsd.Incr("liseth_get_todos_request", []string{"environment:dev"}, 1)
	return func(w http.ResponseWriter, req *http.Request) {
		for _, value := range todoMap {
			fmt.Printf("Todo: %+v\n", value)
		}
	}
}

func postTodo(statsd *statsd.Client) http.HandlerFunc {
	statsd.Incr("liseth_put_request", []string{"environment:dev"}, 1)
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			if err := req.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() error: %v", err)
				return
			}

			fmt.Fprintf(w, "Post from CURL! req.PostForm = %v\n", req.PostForm)
			todoTitle := req.FormValue("todo")
			fmt.Fprintf(w, "TODO: %s\n", todoTitle)

			// Create a new todo and add it to the todo map
			if len(todoTitle) > 0 {
				todo := Todo{title: todoTitle, date: getNowDateTime(), idNum: idNums}
				todoMap[idNums] = todo
				idNums++

				fmt.Println("Current TODOS:")
				for _, value := range todoMap {
					fmt.Printf("Todo: %+v\n", value)
				}
			}
		} else {
			fmt.Printf("Incorrect type of request!\n")
		}
	}
}

// Remove the todo with id from the map, do nothing if it doesn't exist.
func removeTodo(statsd *statsd.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Parse the id from the request
		if req.Method == "POST" {
			if err := req.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() error: %v", err)
				return
			}

			id := req.FormValue("todoId")
			idInt, err := strconv.Atoi(id)
			if err != nil {
				fmt.Println("parsed ID: " + id)
				log.Warn("Error parsing todo id")
				return
			}

			delete(todoMap, idInt)

			fmt.Println("TODOs after delete")
			for _, value := range todoMap {
				fmt.Printf("Todo: %+v\n", value)
			}
		}
	}
}

func main() {
	statsd, err := statsd.New("127.0.0.1:8125")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/hello", hello(statsd))
	http.HandleFunc("/headers", getHeaders(statsd))
	http.HandleFunc("/getTodos", getTodos(statsd))
	http.HandleFunc("/postTodo", postTodo(statsd))
	http.HandleFunc("/removeTodo", removeTodo(statsd))

	todo := Todo{title: "my first todo", date: getNowDateTime(), idNum: idNums}
	idNums++

	// store the TODOs in a slice
	todoMap[idNums] = todo

	http.ListenAndServe(":8090", nil)
}
