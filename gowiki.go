package main

import (
	"html/template"
	"database/sql"
	"log"
	"net/http"
	"os"
	"regexp"
	"fmt"

    "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Page struct {
	PageID  int64
	Title   string
	Body    []byte
}


func (p *Page) save() error {
	pages, err := pagesByTitle(p.Title)
	if err != nil {
		return err
	}
    pg, err := pageByID(pages[0].PageID)
	if err != nil {
		return err
    }
    result, err := db.Exec("UPDATE pages SET Title = ?, Body = ? where PageID = ?", p.Title, p.Body, pg.PageID)
	if err != nil {
		return err
	}
    fmt.Println(result)
	return nil
}

func loadPage(title string) (*Page, error) {
	pages, err := pagesByTitle(title)
	if err != nil {
		return nil, err
	}
    pg, err := pageByID(pages[0].PageID)
	if err != nil {
		return nil, err
	}
    return &Page{Title: pg.Title, Body: pg.Body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+ title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+ title, http.StatusFound)
}

func searchHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
    renderTemplate(w, "search", p)
}

func searchedHandler(w http.ResponseWriter, r *http.Request, title string) {
    title = r.FormValue("body1")
    body := r.FormValue("body2")
    if title != "" {
        pages, err := pagesByTitle(title)
        fmt.Println(err)
        pg, err := pageByID(pages[0].PageID)
        fmt.Println(err)
        http.Redirect(w, r, "/view/"+ pg.Title, http.StatusFound)
    } else {
        pages, err := pagesByBody(body)
        fmt.Println(err)
        pg, err := pageByID(pages[0].PageID)
        fmt.Println(err)
        http.Redirect(w, r, "/view/"+ pg.Title, http.StatusFound)
    }
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html", "search.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view|search|searched)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	// Capture connection properties.
	cfg := mysql.Config{
        User:   os.Getenv("DBUSER"),
        Passwd: os.Getenv("DBPASS"),
        Net:    "tcp",
        Addr:   "127.0.0.1:3306",
        DBName: "articles",
    }
    // Get a database handle.
    var err error
    db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
        log.Fatal(err)
    }

    pingErr := db.Ping()
    if pingErr != nil {
        log.Fatal(pingErr)
    }
    fmt.Println("Connected!")

    http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/search/", makeHandler(searchHandler))
    http.HandleFunc("/searched/", makeHandler(searchedHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// pagesByTitle queries for pages that have the specified title.
func pagesByTitle(title string) ([]Page, error) {
    // A page slice to hold data from returned rows.
    var pages []Page

    rows, err := db.Query("SELECT * FROM pages WHERE Title = ?", title)
    if err != nil {
        return nil, fmt.Errorf("pagesByTitle %q: %v", title, err)
    }
    defer rows.Close()
    // Loop through rows, using Scan to assign column data to struct fields.
    for rows.Next() {
        var pg Page
        if err := rows.Scan(&pg.PageID, &pg.Title, &pg.Body); err != nil {
            return nil, fmt.Errorf("pagesByTitle %q: %v", title, err)
        }
        pages = append(pages, pg)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("pagesByTitle %q: %v", title, err)
    }
    if pages == nil {
        err := addPage(Page{
            Title: title,
            Body: []byte("temp"),
        })
        if err != nil {
            return nil, err
        }
        pagesNew, err := pagesByTitle(title)
        return pagesNew, err
    }
    return pages, nil
}

// pagesByBody queries for pages that have the given body in their own body.
func pagesByBody(body string) ([]Page, error) {
    // A page slice to hold data from returned rows.
    var pages []Page

    rows, err := db.Query("SELECT * FROM pages WHERE Body LIKE ?", "%" + body + "%")
    if err != nil {
        return nil, fmt.Errorf("pagesByBody %q: %v", body, err)
    }
    defer rows.Close()
    // Loop through rows, using Scan to assign column data to struct fields.
    for rows.Next() {
        var pg Page
        if err := rows.Scan(&pg.PageID, &pg.Title, &pg.Body); err != nil {
            return nil, fmt.Errorf("pagesByBody %q: %v", body, err)
        }
        pages = append(pages, pg)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("pagesByBody %q: %v", body, err)
    }
    if pages == nil {
        return nil, fmt.Errorf("pagesByBody %q: %v", body, err)
    }
    return pages, nil
}

// pageByID queries for the page with the specified ID.
func pageByID(id int64) (Page, error) {
    // A page to hold data from the returned row.
    var pg Page

    row := db.QueryRow("SELECT * FROM pages WHERE PageID = ?", id)
    if err := row.Scan(&pg.PageID, &pg.Title, &pg.Body); err != nil {
        if err == sql.ErrNoRows {
            err := addPage(Page{
				Title: pg.Title,
				Body: pg.Body,
			})
			if err != nil {
				return pg, err
			}
            return pg, fmt.Errorf("pageByID %d: no such page", id)
        }
        return pg, fmt.Errorf("pageById %d: %v", id, err)
    }
    return pg, nil
}

// addPage adds the specified page to the database,
// returning the page ID of the new entry
func addPage(pg Page) (error) {
    result, err := db.Exec("INSERT INTO pages (Title, Body) VALUES (?, ?)", pg.Title, pg.Body)
    if err != nil {
        return fmt.Errorf("addPage: %v", err)
    }
    id, err := result.LastInsertId()
	fmt.Println(id)
    if err != nil {
        return fmt.Errorf("addPage: %v", err)
    }
    return nil
}