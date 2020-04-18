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
            }
            .container {
                display: grid;
                margin: 0 150px;
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
                display: inline-block;
                width: 120px;
                padding: 10px 0;
                text-align: center;
                text-decoration: none;
                font-size: 18px;
                border-radius: 6px;
                background-color: #FAD513;
                color: #303030;
            }
            #new_button:hover{
                opacity:0.7;
            }
        </style>
    </head>
    <body>
    <h1>Share Morning Routine</h1>
    <a href="/new" id="new_button">追加</a>
    <br>
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
		//tmpl := template.Must(template.ParseFiles("templates/page.html"))
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
