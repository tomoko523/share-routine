package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/firestore"

	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

const Environment = "prd"

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
        <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css">
        <link href="https://fonts.googleapis.com/css2?family=Dosis:wght@500&display=swap" rel="stylesheet">
        <link href="https://fonts.googleapis.com/css?family=Sawarabi+Gothic" rel="stylesheet">
        <style>
            body {
                background-color: #FFFCDC;
                color: #303030;
                margin: 0 20px;
                font-family: 'Dosis', 'Sawarabi Gothic', sans-serif;
            }
            h1 {
                font-size: 36px;
                text-align: center;
                margin-top: 50px;
            }
            .container {
                display: grid;
                margin: 60px 150px;
                grid-auto-flow: row;
                grid-template-columns: repeat(3, 1fr);
                grid-gap: 20px;
            }
            .card {
                background-image: url(https://storage.googleapis.com/share_routine_cards/mizutama_card.jpg);
                background-size: cover;
                box-shadow: 0 2px 4px rgba(3,3,3,.09);
                border-radius: 3px;
            }
            .card:hover{
                box-shadow: 0 4px 8px rgba(0,0,0,.12);
                border-radius: 3px;
                margin-top:-3px;
            }
            .card a {
                text-decoration: none;
                color: #303030;
            }
            .card a .routines {
                padding-bottom: 40px;
            }
            .user {
                margin: 25px 28px
            }
            .card a .routines .routine {
                margin: 10px 25px;
                padding: 8px 10px;
                background-color: #FFF8A3;
                border-radius: 6px;
            }
            #new_button {
                float: right;
                width: 120px;
            }
        </style>
    </head>
    <body>
    <h1>Share Morning Routine</h1>
    <a href="/new" id="new_button" class="btn btn-warning">追加</a>
    <div class="container">
    {{range $user := .Users}}
        <div class="card">
        <a href="/page/{{$user.ID}}">
            <p class="user">{{$user.Routine.Title}}</p>
            <div class="routines">
                <div class="routine">{{$user.Routine.FirstRoutine}}</div>
                <div class="routine">{{$user.Routine.SecondRoutine}}</div>
                <div class="routine">{{$user.Routine.ThirdRoutine}}</div>
            </div>
        </a>
        </div>
    {{end}}
    </div>
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
        <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css">
		<link href="https://fonts.googleapis.com/css2?family=Dosis:wght@500&display=swap" rel="stylesheet">
        <link href="https://fonts.googleapis.com/css?family=Sawarabi+Gothic" rel="stylesheet">
        <style>
            body {
                background-color: #FFFCDC;
                color: #303030;
                margin: 0 20px;
                font-family: 'Dosis', 'Sawarabi Gothic', sans-serif;
            }
            h1 {
                font-size: 24px;
                text-align: center;
                margin-top: 50px;
            }
            .container {
                width: 70%;
                margin-top: 50px;
            }
            .form-group {
                margin-bottom: 2.5rem;
            }
            input + input {
                margin: 12px 0px;
            }
            .submit-button {
                text-align: center;
            }
            button {
                width: 120px;
            }
            .back-link {
                text-align: center;
                display: block;
                margin-top: 30px;
            }
        </style>
    </head>
    <body>
    <h1>あなたのMorning Routineを登録してね</h1>
    <div class="container">
        <form action="/create" method="post">
            <div class="form-group">
                <label for="title">タイトル</label>
                <input type="text" class="form-control" name="title" id="title" placeholder="tomokoのモーニングルーティン" required>
            </div>
            <div class="form-group">
                <label for="firstRoutine">ルーティーン３つ</label>
                <input type="text" class="form-control" name="firstRoutine" id="firstRoutine" placeholder="ストレッチ、朝ごはん作り、瞑想 etc" required>
                <input type="text" class="form-control" name="secondRoutine" id="secondRoutine" required>
                <input type="text" class="form-control" name="thirdRoutine" id="thirdRoutine" required>
            </div>
            <div class="form-group">
                <label for="message">ひとこと</label>
                <textarea class="form-control" id="message" name="message" rows="3" placeholder="ルーティーンのコンセプトetc" required></textarea>
            </div>
            <div class="submit-button">
                <button class="btn btn-warning">登録</button>
            </div>
    </form>
        <a href="/" class="back-link">Topに戻る</a>
    </div>
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
        <title>{{.User.Routine.Title}}</title>
        <link href="https://fonts.googleapis.com/css2?family=Dosis:wght@500&display=swap" rel="stylesheet">
        <link href="https://fonts.googleapis.com/css?family=Sawarabi+Gothic" rel="stylesheet">
        <style>
            body {
                background-image: url(https://storage.googleapis.com/share_routine_cards/mizutama_1000.jpg);
                background-repeat: repeat;
                color: #303030;
                margin: 0 20px;
                font-family: 'Dosis', 'Sawarabi Gothic', sans-serif;
            }
            .container {
                margin: 60px 150px;
            }
            .container .card {
               margin: 0 10%;
            }
            .title {
                font-size: 38px;
            }
            .routines {

            }
            .routines .routine {
                font-size: 36px;
                margin: 3rem 0;
                padding: 16px 10px;
                background-color: #FFF8A3;
                border-radius: 6px;
            }
            .message {
                border: #FFF9B1 solid 1px;
                background-color: #FFFFFF;
            }
            .message .title {
                text-align: center;
                font-size: 30px;
                color:#FAD513;
                margin-top:2rem;
            }
            .message .main {
                font-size: 28px;
                color: gray;
                margin: 2rem 2.5rem;
            }
            .back-link {
                text-align: center;
                display: block;
                margin-top: 30px;
            }
            .share-link {
                text-align: center;
                display: block;
                margin-top: 30px;
            }
        </style>
    </head>
    <body>
    <div class="container">
        <div class="card">
            <p class="title">{{.User.Routine.Title}}</p>
            <div class="routines">
                <div class="routine">{{.User.Routine.FirstRoutine}}</div>
                <div class="routine">{{.User.Routine.SecondRoutine}}</div>
                <div class="routine">{{.User.Routine.ThirdRoutine}}</div>
            </div>
            <div class="message">
                <div class="title">ひとこと</div>
                <div class="main">{{.User.Routine.Message}}</div>
            </div>
            <a href="{{.User.ShareURL}}" class="share-link">
                <img src="https://storage.googleapis.com/share_routine_cards/twitter_share_button.png">
            </a>
            <a href="/" class="back-link">Topに戻る</a>
        </div>
    </div>
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
	ID       string       `json:"id"`
	Routine  *UserRoutine `json:"routine"`
	ShareURL string       `json:"share_url"`
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

var dummyUserRoutine = UserRoutine{
	Title:         "tomoko",
	FirstRoutine:  "ストレッチ",
	SecondRoutine: "朝ごはん作り",
	ThirdRoutine:  "yoga",
	Message:       "今日はいつもより早起きしてコーヒー淹れたり朝ごはんを丁寧に作ってみました！\nお洋服は新しく買った洋服着てお家勤務もテンション上げて頑張ります〜",
}

var dummyTopPageData = TopPageData{
	Title:   "Share My Routine",
	Message: "ほげほげ",
}

var dummyPageData = PageData{
	User: &User{
		ID:      "5CB42r6DJ0FiD0aGxn3V",
		Routine: &dummyUserRoutine,
	},
	Title:   "Share My Routine",
	Message: "ほげほげ",
}

func TopHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("TopHandler request.")

	tmpl := template.Must(template.New("top").Parse(topTmpl))

	if Environment == "local" {
		tmpl := template.Must(template.ParseFiles("templates/top.html"))
		var dummyUsers []*User
		dummyUsers = append(dummyUsers, &User{
			ID:      "5CB42r6DJ0FiD0aGxn3V",
			Routine: &dummyUserRoutine,
		})
		dummyUsers = append(dummyUsers, &User{
			ID:      "SvRJL7uz13MEwwLvLRql",
			Routine: &dummyUserRoutine,
		})
		dummyUsers = append(dummyUsers, &User{
			ID:      "UAlH3pumAxGChuOuE7Ag",
			Routine: &dummyUserRoutine,
		})
		dummyUsers = append(dummyUsers, &User{
			ID:      "fugagfuga",
			Routine: &dummyUserRoutine,
		})
		dummyUsers = append(dummyUsers, &User{
			ID:      "hogehoge",
			Routine: &dummyUserRoutine,
		})
		dummyTopPageData.Users = dummyUsers
		if err := tmpl.ExecuteTemplate(w, "top", dummyTopPageData); err != nil {
			log.Fatal(err)
		}
		return
	}

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

	if Environment == "local" {
		tmpl = template.Must(template.ParseFiles("templates/newpage.html"))
	}
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

	if Environment == "local" {
		tmpl := template.Must(template.ParseFiles("templates/page.html"))
		// URLを設定
		dummyPageData.setShareURL(r.Host)
		if err := tmpl.ExecuteTemplate(w, "page", dummyPageData); err != nil {
			log.Fatal(err)
		}
		return
	}

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
		Title: res["title"].(string),
	}
	// URLを設定
	data.setShareURL(r.Host)
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

func (data PageData) setShareURL(host string) {
	// シェア用のURLを生成
	v := url.Values{}
	v.Set("text", "私のMorningRoutineは...")
	v.Add("url", fmt.Sprintf("https://%v/page/%v", host, data.User.ID))
	v.Add("hashtags",
		fmt.Sprintf("%v,%v,%v",
			data.User.Routine.FirstRoutine,
			data.User.Routine.SecondRoutine,
			data.User.Routine.ThirdRoutine,
		))
	data.User.ShareURL = fmt.Sprintf("%v%v", "https://twitter.com/intent/tweet?", v.Encode())
}

func main() {
	log.Print("Share MorningRoutine started.")

	r := mux.NewRouter()
	// GET
	r.HandleFunc("/", TopHandler)
	r.HandleFunc("/page/{id}", PageHandler)
	r.HandleFunc("/new", NewPageHandler)
	// POST
	r.HandleFunc("/create", CreateHandler)

	http.ListenAndServe(":8080", r)
}
