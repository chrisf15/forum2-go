package frontend

import (
	"fmt"
	"forum/backend"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

// https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/06.1.html
// read this

// similar to database one

type Drum struct {
	Base *backend.Base
}

// build homepage last? Need to figure out posts and session first
// start with registration?
// func (drum *Drum) Home(w http.ResponseWriter, r *http.Request) {
// 	temp, err := template.ParseFiles("templates/homepage.html")
// 	if err != nil {
// 		http.Error(w, "500 internal service error 500", http.StatusInternalServerError)
// 		return
// 	}
// 	// set up an internal struct to pass and pull info to.
// 	// make them interfaces to store more info
// 	type basedata struct {
// 		cookies interface{}
// 		posts   interface{}
// 	}
// 	// then use a variable to store them
// 	var page basedata
// 	sortP := r.FormValue("sort")
// 	cky, err := r.Cookie("sessionIdentity")
// }

// https://stackoverflow.com/questions/27234861/correct-way-of-getting-clients-ip-addresses-from-http-request
// https://golangcode.com/get-the-request-ip-addr/ this one is smaller change var names
func FindIP(r *http.Request) string {
	address := r.Header.Get("X-FORWARDED-FOR")
	if address != "" {
		return address
	}
	// RemoteAddr allows HTTP servers and other software to record
	// the network address that sent the request, usually for
	// logging. This field is not filled in by ReadRequest and
	// has no defined format. The HTTP server in this package
	// sets RemoteAddr to an "IP:port" address before invoking a
	// handler.
	// This field is ignored by the HTTP client.
	// RemoteAddr string
	return r.RemoteAddr
}

func (drum *Drum) StartPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}
	// like mo built but more cases?
	// handle the back end in a different function
	switch r.Method {
	default:
		http.Error(w, "400 Bad Request 400", http.StatusBadRequest)
	case "POST":
		// username := r.FormValue("username")
		// password := r.FormValue("password")
		// email := r.FormValue("email")
		// if username == "" || password == "" || email == "" {
		// 	http.Error(w, "400 Bad Request 400", http.StatusBadRequest)
		// 	return
		// }
		// // call in other functions recursively. USerAgent = Identity.
		// // https://golangbyexample.com/user-agent-http-golang/
		// _, _, _, err := drum.Base.Register(username, email, r.UserAgent(), FindIP(r), password)
		// if err != nil {
		// 	w.WriteHeader(http.StatusOK)
		// 	w.Header().Set("Content-type", "application/text")
		// 	w.Write([]byte("0" + err.Error()))
		// 	return
		// }
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "application/text")
		// https://stackoverflow.com/questions/37863374/whats-the-difference-between-responsewriter-write-and-io-writestring
		w.Write([]byte("Registered successfully - You may now log in"))
	case "GET":
		files := GetTemplates()
		RenderTemplate(w, r, files, "startpage", "")
	}
}

func (drum *Drum) Homepage(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Cookies interface{}
		Posts   interface{}
	}
	var pagePres pageData
	// i need some front end please
	// if r.URL.Path != "/homepage" {
	// 	http.Error(w, "404 page not found 404", http.StatusNotFound)
	// 	return
	// }
	web, err := template.ParseFiles("templates/homepage.html")
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	// need a sort func
	// can it be built in?
	sortBy := r.FormValue("sortBy")
	//c, err := r.Cookie("session")
	if err != nil {
		pagePres = pageData{
			Cookies: err.Error(),
			Posts:   drum.Base.PostIndex(sortBy, ""),
		}
		if err := web.Execute(w, pagePres); err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
		// } else {
		// 	validCookie := drum.IsCookieValid(w, c)
		// 	IDPost := r.FormValue("postID")

		// 	commentBody := r.FormValue("comment")
		// 	// need func to update/make the reactions

		// 	upvoteCom := r.FormValue("upvoteCom")
		// 	downvoteCom := r.FormValue("upvoteCom")
		// 	upvotePost := r.FormValue("upvotePost")
		// 	downvotePost := r.FormValue("downvotePost")

		// 	if commentBody != "" {
		// 		drum.Base.CommentComment(validCookie[0], IDPost, commentBody)
		// 	}
		// 	if upvoteCom != "" {
		// 		// cU := strings.Split(upvoteCom, "&")
		// 		// drum.Base.UpdateCommentReaction(cU[1], cU[2], validCookie[0], cU[0])
		// 	}
		// 	if downvoteCom != "" {
		// 		// cD := strings.Split(downvoteCom, "&")
		// 		// drum.Base.UpdateCommentReaction(cD[1], cD[2], validCookie[0], cD[0])
		// 	}
		// 	if upvotePost != "" {
		// 		pU := strings.Split(upvotePost, "&")
		// 		drum.Base.UpdatePostReaction(pU[1], validCookie[0], pU[0])
		// 	}
		// 	if downvotePost != "" {
		// 		pD := strings.Split(downvotePost, "&")
		// 		drum.Base.UpdatePostReaction(pD[1], validCookie[0], pD[0])
		// 	}

	}
	files := GetTemplates()
	RenderTemplate(w, r, files, "Homepage", drum.Base.PostIndex(sortBy, ""))
}

// duplicated this to find a way to show comments on each post
// will have to sort by the postID. can be done in the html?

func (drum *Drum) PostComments(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Cookies  interface{}
		Comments interface{}
	}
	var pagePres pageData
	// i need some front end please
	// if r.URL.Path != "/homepage" {
	// 	http.Error(w, "404 page not found 404", http.StatusNotFound)
	// 	return
	// }
	web, err := template.ParseFiles("templates/comments.html")
	if err != nil {
		http.Error(w, "500 Internal Server Error1", http.StatusInternalServerError)
		return
	}
	// need a sort func
	// can it be built in?
	// sortBy := r.FormValue("sortBy")
	c, err := r.Cookie("session")
	if err != nil {
		pagePres = pageData{
			Cookies: err.Error(),
			// Comments: drum.Base.CommentIndex(""),
		}
		if err := web.Execute(w, pagePres); err != nil {
			http.Error(w, "500 Internal Server Error2", http.StatusInternalServerError)
			return
		}
	} else {
		validCookie := drum.IsCookieValid(w, c)
		IDPost := r.FormValue("postID")

		commentBody := r.FormValue("comment")
		// need func to update/make the reactions

		// upvoteCom := r.FormValue("upvoteCom")
		// downvoteCom := r.FormValue("upvoteCom")
		// upvotePost := r.FormValue("upvotePost")
		// downvotePost := r.FormValue("downvotePost")

		if commentBody != "" {
			drum.Base.CommentComment(validCookie[0], IDPost, commentBody)
		}
	}
	// 	if upvoteCom != "" {
	// 		cU := strings.Split(upvoteCom, "&")
	// 		drum.Base.UpdateCommentReaction(cU[1], cU[2], validCookie[0], cU[0])
	// 	}
	// 	if downvoteCom != "" {
	// 		cD := strings.Split(downvoteCom, "&")
	// 		drum.Base.UpdateCommentReaction(cD[1], cD[2], validCookie[0], cD[0])
	// 	}
	// 	if upvotePost != "" {
	// 		pU := strings.Split(upvotePost, "&")
	// 		drum.Base.UpdatePostReaction(pU[1], validCookie[0], pU[0])
	// 	}
	// 	if downvotePost != "" {
	// 		pD := strings.Split(downvotePost, "&")
	// 		drum.Base.UpdatePostReaction(pD[1], validCookie[0], pD[0])
	// 	}

	// }
	files := GetTemplates()
	RenderTemplate(w, r, files, "Comments", drum.Base.CommentIndex(""))
}

// could handle first bit with a separate func rather than calling into all of them?
func (drum *Drum) Register(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/register" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}
	// like mo built but more cases?
	// handle the back end in a different function
	switch r.Method {
	default:
		http.Error(w, "400 Bad Request 400", http.StatusBadRequest)
	case "GET":
		files := GetTemplates()
		RenderTemplate(w, r, files, "Register", "")
	case "POST":
		username := r.FormValue("username")
		password := r.FormValue("password")
		email := r.FormValue("email")
		if username == "" || password == "" || email == "" {
			http.Error(w, "400 Bad Request 400", http.StatusBadRequest)
			return
		}
		// call in other functions recursively. USerAgent = Identity.
		// https://golangbyexample.com/user-agent-http-golang/
		_, _, _, err := drum.Base.Register(username, email, password)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-type", "application/text")
			w.Write([]byte("0" + err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "application/text")
		// https://stackoverflow.com/questions/37863374/whats-the-difference-between-responsewriter-write-and-io-writestring
		// w.Write([]byte("Registered successfully - You may now log in"))
		// http.Redirect(w, r, "/", http.StatusSeeOther)
		files := GetTemplates()
		RenderTemplate(w, r, files, "Register", "Registered successfully - You may now log in")
	}
}

// https://golang.hotexamples.com/examples/net.http/Cookie/Expires/golang-cookie-expires-method-examples.html
// third example is good for logging in
func (drum *Drum) MyCrewIsLoggingOn(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}
	// r.UserAgent returns clients user agent/identity
	// identity := r.UserAgent()
	// next bit like what Mo built but more cases?
	switch r.Method {
	default:
		http.Error(w, "400 bad boy request 400", http.StatusBadRequest)
	case "POST":
		nameUser := r.FormValue("username")
		password := r.FormValue("password")
		if nameUser == "" || password == "" {
			http.Error(w, "400 bad request 400", http.StatusBadRequest)
			return
		}
		_, _, _, err := drum.Base.LoginUser(nameUser, password)
		if err != nil {
			// https://golangbyexample.com/set-resposne-headers-http-go/
			// above link has a spelling mistake
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/text")
			// https://stackoverflow.com/questions/43470284/bytestring-vs-bytestring/43470344#43470344
			w.Write([]byte("0" + err.Error()))
		}
		// initialise the cookie here
		// https://golangbyexample.com/set-cookie-http-golang/
		// set an expiry?
		fmt.Print("")

		sID := uuid.NewV4()
		c := &http.Cookie{
			Name: "session",
			// this value makes it unique to the user
			Value: sID.String(),
			// this means it can be used globallly?
			Path: "/",
		}
		c.MaxAge = 30000
		http.SetCookie(w, c)
		backend.SessionDB[c.Value] = backend.Session{}
		http.Redirect(w, r, "http://localhost:8080", http.StatusSeeOther)
		return
		// w.WriteHeader(http.StatusOK)
		// w.Header().Set("Content-Type", "application/text")
		// use Write[]byte to change the info put in into bytes.
		// probs best way to do it without running into errors
		// as several different bits of info not all strings
		// w.Write([]byte("1" + userID + "-" + sessionID + "-" + username))
	case "GET":
		files := GetTemplates()
		RenderTemplate(w, r, files, "LogIn", "")
	}
}

// similar to log in but without the formParsing
// use the cookie applied to identify?
func (drum *Drum) LogOut(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/logout" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}
	var sesId string
	if drum.Base.IsSessionValid(sesId) {
		http.Redirect(w, r, "/homepage", http.StatusSeeOther)
		return
	}
	switch r.Method {
	default:
		http.Error(w, "400 bad request 400", http.StatusBadRequest)
		return
	case "GET":
		files := GetTemplates()
		RenderTemplate(w, r, files, "LogIn", "")
	case "POST":
		cky, err := r.Cookie("session")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp := strings.Split(cky.Value, "&")
		err = drum.Base.DeleteSession(resp[1]) // maybe needs more info here [1] as it wants the first argument/it is a []string
		if err != nil {
			log.Fatal(err)
		}
		// create a new token for the unlogged user
		http.SetCookie(w, &http.Cookie{
			Name:    "session",
			Value:   "",
			Expires: time.Now(),
		})
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("Logged out succesfully"))
	}
}

// similar to the others as well
//"if it aint broke don't fix it"
// post and comment almost exactly the same just talk to slightly different bits

func (drum *Drum) MakePost(w http.ResponseWriter, r *http.Request) {
	//var user backend.User
	if r.URL.Path != "/post" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}

	// if user.LoggedIn != true {
	// 	http.Redirect(w, r, "/homepage", http.StatusSeeOther)
	// 	return
	// }
	//var sessId string
	// if foundSession, sessionValue := drum.Base.IsSessionValid(sessId); foundSession {
	// 	_, found := drum.Base.GetUser(sessionValue)
	// 	if !found {
	// 		http.Redirect(w, r, "/homepage", http.StatusSeeOther)
	// 		return
	// 	}
	// }

	// cky, _ := r.Cookie("session")
	// need to check cookies validity, new func of course :(
	// cook := drum.IsCookieValid(w, cky)
	// if err == nil {
	switch r.Method {
	default:
		http.Error(w, "400 Bad Boi Requested 400", http.StatusBadRequest)
	case "GET":
		files := GetTemplates()
		RenderTemplate(w, r, files, "Post", "")
	case "POST":
		title := r.FormValue("title")

		category := r.FormValue("category")
		category2 := r.FormValue("category2")
		post := r.FormValue("body")
		_, err := drum.Base.PostPost("", title, category, category2, post)
		if err != nil {
			fmt.Println(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "application/text")
		// postID, _ := drum.Base.PostPost(cook[0], title, category, category2, post)
		// w.WriteHeader(http.StatusOK)
		// w.Header().Set("Content-Type", "application/text")
		// w.Write([]byte(postID))
		files := GetTemplates()
		RenderTemplate(w, r, files, "Post", "Posted successfully - You may now comment")
	}
	//}
}

func (drum *Drum) WriteComment(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/comment" {
		http.Error(w, "404 page not found 404", http.StatusNotFound)
		return
	}
	//backend.GetUser(w, r)
	//u := backend.GetUser(w, r)
	var sesId string
	if drum.Base.IsSessionValid(sesId) {
		http.Redirect(w, r, "/homepage", http.StatusSeeOther)
		return
	}
	// cky, err := r.Cookie("session")
	// // need to check cookies validity, new func of course :(
	// cook := drum.IsCookieValid(w, cky)
	// if err == nil {
	switch r.Method {
	default:
		http.Error(w, "400 Bad Boi Requested 400", http.StatusBadRequest)
	case "GET":
		files := GetTemplates()
		RenderTemplate(w, r, files, "Comment", "")
	case "POST":
		body := r.FormValue("body")
		//postID := drum.Base.DB.Query("SELECT")
		/*commentID*/
		_, err := drum.Base.CommentComment( /*cook[0]*/ "22" /*postID*/, "", body)
		if err != nil {
			fmt.Println(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/text")
		// w.Write([]byte(commentID))
		// }
		files := GetTemplates()
		RenderTemplate(w, r, files, "Comment", "Thanks for contributing to the discussion")
	}
}

// has to be a new func as you can do it using http.SetCookie(something, somethingElse) but I don't think
// you can call that into other funcs
// eg something.Valid is a way to check within a func
// https://github.com/golang/go/issues/46370

func (drum *Drum) IsCookieValid(w http.ResponseWriter, c *http.Cookie) []string {
	cky := []string{}
	if strings.Contains(c.String(), "&") {
		cky = strings.Split(c.Value, "&")
	}
	if len(cky) != 0 {
		// [1] as it takes the first argument
		if !(drum.Base.IsSessionValid(cky[1])) {
			http.SetCookie(w, &http.Cookie{
				Name:    "session",
				Value:   "",
				Expires: time.Now(),
			})
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/text")
			w.Write([]byte("OI BLUD you're not signed in geeeeeezer"))
		} else {
			// the []string in the func title helps this return
			// won't return without it
			return cky
		}
	}
	return cky
}

// funcs that could be used to open and post files?
// would function names need to be different?
// func GetTemplates() []string {
// 	files := []string{}
// 	folder, _ := ioutil.ReadDir("templates")
// 	for _, subitem := range folder {
// 		files = append(files, "./templates/"+subitem.Name())
// 	}

// 	return files
// }

// Render template on get request.
func RenderTemplate(w http.ResponseWriter, r *http.Request, files []string, templateName string, data interface{}) {
	// Parse files, check for errors.
	tmplSet, err := template.ParseFiles(files...)
	if err != nil {
		fmt.Print(err)
	}

	// Execute homepage template. (Send HTML to the front-end.)
	err = tmplSet.ExecuteTemplate(w, templateName, data)
	if err != nil {
		fmt.Print(err)
	}
}

func GetTemplates() []string {
	files := []string{}
	folder, _ := ioutil.ReadDir("templates")
	for _, subitem := range folder {
		files = append(files, "./templates/"+subitem.Name())
	}

	return files
}

// 	// Execute homepage template. (Send HTML to the front-end.)
// 	err = tmplSet.ExecuteTemplate(w, templateName, data)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// if err == http.ErrNoCookie {
// 	// https://go.dev/src/net/http/status.go
// 	// https://golangbyexample.com/401-http-status-response-go/
// 	w.WriteHeader(http.StatusUnauthorized)
