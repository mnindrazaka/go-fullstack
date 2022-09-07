package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

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

			name := r.FormValue("name")
			uploadedFile, handler, err := r.FormFile("photo")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer uploadedFile.Close()

			dir, err := os.Getwd()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			filename := fmt.Sprintf("%s%s", name, filepath.Ext(handler.Filename))
			fileLocation := filepath.Join(dir, "assets", "upload", filename)
			targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer targetFile.Close()

			if _, err := io.Copy(targetFile, uploadedFile); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = db.Exec("INSERT INTO student SET name=?, photo=?", name, filename)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		} else {
			http.Error(w, "", http.StatusBadRequest)
		}
	})

	http.HandleFunc("/edit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			id := r.URL.Query().Get("id")

			student := student{}
			if err := db.QueryRow("SELECT * FROM student WHERE id = ?", id).Scan(&student.Id, &student.Name, &student.Photo); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			tmpl, err := template.ParseFiles("views/edit.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := tmpl.Execute(w, student); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else if r.Method == "POST" {
			if err := r.ParseMultipartForm(1024); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			id := r.URL.Query().Get("id")
			name := r.FormValue("name")

			if _, err := db.Exec("UPDATE student SET name=? WHERE id=?", name, id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			uploadedFile, _, err := r.FormFile("photo")

			switch err {
			case nil:
				defer uploadedFile.Close()
				dir, err := os.Getwd()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				student := student{}
				if err := db.QueryRow("SELECT * FROM student WHERE id = ?", id).Scan(&student.Id, &student.Name, &student.Photo); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				filename := student.Photo
				fileLocation := filepath.Join(dir, "assets", "upload", filename)

				targetFile, err := os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer targetFile.Close()

				if _, err := io.Copy(targetFile, uploadedFile); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/", http.StatusMovedPermanently)
			case http.ErrMissingFile:
				http.Redirect(w, r, "/", http.StatusMovedPermanently)
				break
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
				break
			}
		} else {
			http.Error(w, "", http.StatusBadRequest)
		}
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		student := student{}

		if err := db.QueryRow("SELECT * FROM student WHERE id=?", id).Scan(&student.Id, &student.Name, &student.Photo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		dir, err := os.Getwd()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filename := student.Photo
		fileLocation := filepath.Join(dir, "assets", "upload", filename)
		err = os.Remove(fileLocation)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := db.Exec("DELETE FROM student WHERE id=?", id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusMovedPermanently)
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
