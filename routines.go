package main

import (
	"context"
	"html/template"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

// HTML
var topTmpl = `
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

var newPageTmpl = `
{{define "newpage"}}
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
</head>
<body>
<form action="/create" method="post">
    <div>
        <label for="title">名前をつけてください</label>
        <input type="text" name="title" id="title" placeholder="私のモーニングルーティン" required>
    </div>
    <div>
        <label for="firstRoutine">あなたが起きて最初にすることは?</label>
        <input type="text" name="firstRoutine" id="firstRoutine" placeholder="大きく背を伸びる" required>
    </div>
    <div>
        <label for="secondRoutine">あなたが起きて２番目にすることは?</label>
        <input type="text" name="secondRoutine" id="secondRoutine" placeholder="カーテンを開けて日光をあびる">
    </div>
    <div>
        <label for="thirdRoutine">あなたが起きて３番目にすることは?</label>
        <input type="text" name="thirdRoutine" id="thirdRoutine" placeholder="コーヒーを入れる">
    </div>
    <div>
        <label for="message">みんなにメッセージをどうぞ</label>
        <input type="text" name="message" id="message" placeholder="気持ちのいい朝を過ごしましょう！！" required>
    </div>
    <div>
        <button>登録する</button>
    </div>
</form>
</body>
</html>
{{end}}
`

var pageTmpl = `
{{define "page"}}
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

// UserRoutine ... ルーティン
type UserRoutine struct {
	Title         string `firestore:"title"`
	FirstRoutine  string `firestore:"first_routine"`
	SecondRoutine string `firestore:"second_routine"`
	ThirdRoutine  string `firestore:"third_routine"`
	Message       string `firestore:"message"`
	ImageURL      string `firestore:"image_url"`
}

func TopHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("TopHandler request.")

	tmpl := template.Must(template.New("top").Parse(topTmpl))

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

func NewPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("NewPageHandler request.")

	tmpl := template.Must(template.New("newpage").Parse(newPageTmpl))
	dat := struct {
		Title string
	}{
		Title: "新規作成",
	}

	if err := tmpl.ExecuteTemplate(w, "newpage", dat); err != nil {
		log.Fatal(err)
	}
}

func PageHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("PageHandler request.")

	tmpl := template.Must(template.New("page").Parse(pageTmpl))
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
	if err := tmpl.ExecuteTemplate(w, "page", dat); err != nil {
		log.Fatal(err)
	}
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("CreateHandler request.")

	// 受け取った値を取得する
	err := r.ParseForm()
	if err != nil {
		log.Printf("Parse error : %v", err)
	}

	var routine UserRoutine
	var decoder = schema.NewDecoder()
	err = decoder.Decode(&routine, r.PostForm)
	if err != nil {
		log.Printf("decode error : %v", err)
	}
	log.Println(routine)

	// 登録する
	projectID := "share-my-routine"
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Printf("firebase.NewApp error : %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("Firestore error : %v", err)
	}
	defer client.Close()

	_, _, err = client.Collection("userRoutines").Add(ctx, routine)
	if err != nil {
		log.Printf("Add collection error : %v", err)
		return
	}

	// topに戻す
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
	r.HandleFunc("/new", NewPageHandler)
	// POST
	r.HandleFunc("/create", CreateHandler)

	http.ListenAndServe(":8080", r)
}
