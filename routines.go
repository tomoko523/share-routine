package main

import (
	"context"
	"html/template"
	"log"
	"net/http"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/firestore"

	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

const FireStoreProjectID = "share-my-routine"
const CollectionName = "userRoutines"

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
<a href="/new">追加</a>
<br>
{{range $user := .Users}}
  <p>{{$user.ID}}</p>
  <p>{{$user.Routine.Title}}</p>
  <p>{{$user.Routine.FirstRoutine}}</p>
  <p>{{$user.Routine.SecondRoutine}}</p>
  <p>{{$user.Routine.ThirdRoutine}}</p>
  <p>{{$user.Routine.Message}}</p>
  <p><a href="/page/{{$user.ID}}">{{$user.Routine.Title}}</a></p>
{{end}}
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
<p>{{.User.ID}}</p>
<p>{{.User.Routine.Title}}</p>
<p>{{.User.Routine.FirstRoutine}}</p>
<p>{{.User.Routine.SecondRoutine}}</p>
<p>{{.User.Routine.ThirdRoutine}}</p>
<p>{{.User.Routine.Message}}</p>
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

type User struct {
	ID      string       `json:"id"`
	Routine *UserRoutine `json:"routine"`
}

type TopPageData struct {
	Users   []*User
	Title   string
	Message string
}

type PageData struct {
	User    *User
	Title   string
	Message string
}

func TopHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("TopHandler request.")

	tmpl := template.Must(template.New("top").Parse(topTmpl))

	ctx := context.Background()
	client, err := createFirestoreConnection(ctx, FireStoreProjectID)
	if err != nil {
		log.Printf("Firestore error : %v", err)
	}
	defer client.Close()

	iter := client.Collection("userRoutines").Documents(ctx)
	var users []*User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("GetALL error : %v", err)
			return
		}
		res := doc.Data()
		users = append(users, &User{
			ID: doc.Ref.ID,
			Routine: &UserRoutine{
				Title:         res["title"].(string),
				FirstRoutine:  res["first_routine"].(string),
				SecondRoutine: res["second_routine"].(string),
				ThirdRoutine:  res["third_routine"].(string),
				Message:       res["message"].(string),
			},
		})
	}

	data := TopPageData{
		Users:   users,
		Title:   "Share My Routine",
		Message: "ほげほげ",
	}

	if err := tmpl.ExecuteTemplate(w, "top", data); err != nil {
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
	idStr := vars["id"]

	ctx := context.Background()
	client, err := createFirestoreConnection(ctx, FireStoreProjectID)
	if err != nil {
		log.Printf("Firestore error : %v", err)
	}
	defer client.Close()

	doc, err := client.Collection(CollectionName).Doc(idStr).Get(ctx)
	if err != nil {
		log.Printf("Get by id error : %v", err)
	}
	res := doc.Data()

	data := PageData{
		User: &User{
			ID: idStr,
			Routine: &UserRoutine{
				Title:         res["title"].(string),
				FirstRoutine:  res["first_routine"].(string),
				SecondRoutine: res["second_routine"].(string),
				ThirdRoutine:  res["third_routine"].(string),
				Message:       res["message"].(string),
			},
		},
		Title:   res["title"].(string),
		Message: "ふがふが",
	}

	if err := tmpl.ExecuteTemplate(w, "page", data); err != nil {
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

	// 登録するFireStoreProjectID
	ctx := context.Background()
	client, err := createFirestoreConnection(ctx, FireStoreProjectID)
	if err != nil {
		log.Printf("Firestore error : %v", err)
	}
	defer client.Close()

	_, _, err = client.Collection(CollectionName).Add(ctx, routine)
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

func createFirestoreConnection(ctx context.Context, pID string) (*firestore.Client, error) {
	conf := &firebase.Config{ProjectID: pID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Printf("firebase.NewApp error : %v", err)
	}
	return app.Firestore(ctx)
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
