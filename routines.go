package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func TopHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("TopHandler request.")
	var tmpText = `
{{define "top"}}
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
</head>
<body>
{{.Message}}
</body>
</html>
{{end}}
`
	tmpl := template.Must(template.New("top").Parse(tmpText))

	// TODO: 値を取ってきて一覧表示する

	dat := struct {
		Title   string
		Message string
	}{
		Title:   "Test",
		Message: "ほげほげほげほげ",
	}
	if err := tmpl.ExecuteTemplate(w, "top", dat); err != nil {
		log.Fatal(err)
	}
}

func PageHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("PageHandler request.")
	var tmpText = `
{{define "top"}}
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
</head>
<body>
{{.Message}}
{{.ID}}
</body>
</html>
{{end}}
`

	tmpl := template.Must(template.New("top").Parse(tmpText))
	vars := mux.Vars(r)

	// TODO: 値をとってきてテンプレートに入れる

	dat := struct {
		Title   string
		Message string
		ID      string
	}{
		Title:   "Test",
		Message: "ほげほげほげほげ",
		ID:      vars["id"],
	}
	if err := tmpl.ExecuteTemplate(w, "top", dat); err != nil {
		log.Fatal(err)
	}
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("CreateHandler request.")

	// TODO 受け取った値を取得する
	//err := r.ParseForm()
	//if err != nil {
	//	// Handle error
	//}
	//
	//var person Person
	//// POSTフォームで送信された変数をstructに変換する
	//err = decoder.Decode(&person, r.PostForm)
	//if err != nil {
	//	// Handle error
	//}

	// TODO 値を登録する

	// TODO topにリダイレクト(個人ページに戻す）
	newURL := "https://" + r.Host
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("location", newURL)
	w.WriteHeader(http.StatusMovedPermanently)
}

func main() {
	log.Print("Hello world sample started.")
	r := mux.NewRouter()
	// GET
	r.HandleFunc("/", TopHandler)
	r.HandleFunc("/page/{id}", PageHandler)
	//r.HandleFunc("/new", NewPageHandler)

	// POST
	r.HandleFunc("/create", CreateHandler)

	http.ListenAndServe(":8080", r)
}
