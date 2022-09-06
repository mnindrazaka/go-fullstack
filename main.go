package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type student struct {
	Id    int
	Name  string
	Photo string
}

func main() {
	db, err := sql.Open("mysql", "root:((root))@tcp(127.0.0.1:3306)/school")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("assets"))))

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			tmpl, err := template.ParseFiles("views/create.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = tmpl.Execute(w, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if r.Method == "POST" {
			if err := r.ParseMultipartForm(1024); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		} else {
			http.Error(w, "", http.StatusBadRequest)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM student")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var students []student

		for rows.Next() {
			student := student{}
			rows.Scan(&student.Id, &student.Name, &student.Photo)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			students = append(students, student)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		funcMap := template.FuncMap{
			"plus": func(a int, b int) int {
				return a + b
			},
		}

		tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("views/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"students": students,
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err = http.ListenAndServe("127.0.0.1:3000", nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}
