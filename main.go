package main

import (
    "database/sql"
	//"log"
	"os"
	"fmt"
    "net/http"
    "html/template"
	"strconv"
	

    _ "github.com/go-sql-driver/mysql"
    "github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// You could set this to any `io.Writer` such as a file
  file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
   if err == nil {
	log.SetOutput(file)
  } else {
    log.Info("Failed to log to file, using default stderr")
   }
  }

type customer struct {
    amount  int
	name  string 
	pwd []byte
}


var templates *template.Template
var store = sessions.NewCookieStore([]byte("Ramya"))
var db *sql.DB
var err error

func main() {
    templates = template.Must(template.ParseGlob("templates/*.html"))
    r := mux.NewRouter()
      
	r.HandleFunc("/", gethome).Methods("GET")
      r.HandleFunc("/login", getlogin).Methods("GET")
      r.HandleFunc("/login", postlogin).Methods("POST")
      r.HandleFunc("/deposit", getdeposit).Methods("GET")
      r.HandleFunc("/deposit", postdeposit).Methods("POST")
      r.HandleFunc("/withdraw", getwithdraw).Methods("GET")
      r.HandleFunc("/withdraw", postwithdraw).Methods("POST")
	  r.HandleFunc("/checkbalance", getcheckbalance).Methods("GET")
	  r.HandleFunc("/signup", getregister).Methods("GET")
	  r.HandleFunc("/signup", postregister).Methods("POST")
	  r.HandleFunc("/logout", getlogout).Methods("GET")
      
       r.HandleFunc("/index", getindex).Methods("GET")
       http.Handle("/", r)
      http.ListenAndServe(":8080", nil)
      
  }

  func getlogout(w http.ResponseWriter, r *http.Request){
	

	http.Redirect(w, r, "/", 302)
	log.WithFields(log.Fields{
		}).Info( "logged out Succesfully")
    //templates.ExecuteTemplate(w, "home.html", nil)
 } 
 
  func gethome(w http.ResponseWriter, r *http.Request){
    templates.ExecuteTemplate(w, "home.html", nil)
 } 

 func getlogin(w http.ResponseWriter, r *http.Request){
    templates.ExecuteTemplate(w, "login.html", nil)
 }
  func postlogin(w http.ResponseWriter, r *http.Request) {
      r.ParseForm()
    db, err := sql.Open("mysql", "root:R@mya999@(127.0.0.1:3306)/dbname")
    name1 := r.PostForm.Get("name")
	pwd1 := r.PostForm.Get("pwd")
	Result, err := db.Query("SELECT * FROM customer WHERE name=?", name1)
    user := customer{}
	for Result.Next() {
		var name2 string
		var amount int
		var pwd2 []byte
		err = Result.Scan(&name2, &pwd2, &amount)
		if err != nil {
			panic(err.Error())
		}
		user.pwd = pwd2
	}
    

     hashFromDatabase := user.pwd
	if err := bcrypt.CompareHashAndPassword(hashFromDatabase, []byte(pwd1)); err != nil {
		// TODO: Properly handle error
		templates.ExecuteTemplate(w, "login.html", "invalid login")
		return
     //   log.Fatal(err)
    } else {
		session, _ := store.Get(r, "session")
		session.Values["name"] = name1
		session.Save(r, w)
		http.Redirect(w, r, "/index", 302)
	}
	templates.ExecuteTemplate(w, "login", nil)
	
	log.WithFields(log.Fields{
		"user": name1,
	  }).Info( "logged in Succesfully")
	
	  defer db.Close()
}

func getregister(w http.ResponseWriter, r *http.Request){
	templates.ExecuteTemplate(w, "register.html", nil)
}
func postregister(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
    db, err := sql.Open("mysql", "root:R@mya999@(127.0.0.1:3306)/dbname")
	log.Println("opening register")
	name := r.PostForm.Get("name")
	pwd := r.PostForm.Get("pwd")
	amount := 0

	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)


	_, err = db.Exec("INSERT INTO customer (name, pwd, amount) VALUES (?, ?, ?)", name, hash, amount)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"user": name,
	  }).Info( "Registered")



	http.Redirect(w, r, "/login", 301)
	templates.ExecuteTemplate(w, "register.html", nil)
	defer db.Close()
}

func getdeposit(w http.ResponseWriter, r *http.Request) {
    templates.ExecuteTemplate(w, "deposit.html", nil)
}

func postdeposit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
    db, err := sql.Open("mysql", "root:R@mya999@(127.0.0.1:3306)/dbname?parseTime=true")
    session, _ := store.Get(r, "session")
	name1, ok := session.Values["name"]
	if !ok {
		http.Redirect(w, r, "/login", 302)
		return
    }
    amount := r.PostForm.Get("amount")
	Result, err := db.Query("SELECT * FROM customer WHERE name=?", name1)
    user := customer{}
	for Result.Next() {
		var name2, pwd2 string
		var amount2 int
		err = Result.Scan(&name2, &pwd2, &amount2)
		if err != nil {
			panic(err.Error())
		}
		user.amount = amount2
    }
    amount1, err := strconv.Atoi(amount)
	if err != nil {
		fmt.Println("Enter valid amount")
	}
	amount1 = amount1 + user.amount
	_, err = db.Exec("UPDATE customer SET amount=? WHERE name=?", amount1, name1)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"amount": amount,
	  }).Info( "amount deposited")
	
	http.Redirect(w, r, "/index", 302)
	defer db.Close()
}

func getwithdraw(w http.ResponseWriter, r *http.Request){
    templates.ExecuteTemplate(w, "withdraw.html", nil)
}
func postwithdraw(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
    db, err := sql.Open("mysql", "root:R@mya999@(127.0.0.1:3306)/dbname?parseTime=true")
    session, _ := store.Get(r, "session")
	name1, ok := session.Values["name"]
	if !ok {
		http.Redirect(w, r, "/login", 302)
		return
    }
    amount := r.PostForm.Get("amount")
	Result, err := db.Query("SELECT * FROM customer WHERE name=?", name1)
    user := customer{}
	for Result.Next() {
		var name2, pwd2 string
		var amount2 int
		err = Result.Scan(&name2, &pwd2, &amount2)
		if err != nil {
			panic(err.Error())
		}
		user.amount = amount2
    }
    amount1, err := strconv.Atoi(amount)
	if err != nil {
		fmt.Println("Enter valid amount")
	}
	amount1 = user.amount - amount1 
	_, err = db.Exec("UPDATE customer SET amount=? WHERE name=?", amount1, name1)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"amount": amount,
	  }).Info( "amount withdrawed")

	http.Redirect(w, r, "/index", 302)
	defer db.Close()
}

func getcheckbalance(w http.ResponseWriter, r *http.Request) {
    db, err := sql.Open("mysql", "root:R@mya999@(127.0.0.1:3306)/dbname")
	session, _ := store.Get(r, "session")
	name, ok := session.Values["name"]
	if !ok {
		http.Redirect(w, r, "/login", 302)
		return
	}
    Result, err := db.Query("SELECT * FROM customer WHERE name=?", name)
	user := customer{}
	for Result.Next() {
		var name2, pwd2 string
		var amount2  int
		err = Result.Scan(&name2, &pwd2, &amount2)
		if err != nil {
			panic(err.Error())
		}
        user.amount = amount2
    }
    amount3 := user.amount
	templates.ExecuteTemplate(w, "check.html", amount3)

	log.WithFields(log.Fields{
		"amount": amount3,
	  }).Info( "Balance check")

	defer db.Close()
}

func getindex(w http.ResponseWriter, r *http.Request) {
    templates.ExecuteTemplate(w, "index.html", nil)
}
