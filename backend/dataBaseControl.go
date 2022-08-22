package backend

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// a lot of the functions here end up being the same as its just
// inserting the info into different databases,
// will then need to add other functions to handle the rest
// eg update and delete

// UUID seems easier to use/no harder than learning the packages within go
// call NewV4 as it adds a random Unique identifier
// https://pkg.go.dev/github.com/satori/go.uuid

// doesnt allow this as a constant, is there another way?
// const date = time.Now().Format("2004.04.20 04:20:00")

/* only use (base *Base) and not (base *Base, user *User) as there can only be
one reciever in each function. Could get round it and simplify code if using a
JSON as well as a SQL database
*/

// HashPassword turns the password into a hashed string
func CreateHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}

// CheckPasswordHash checks the entered password against the hashed password
func CheckHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (base *Base) GetUser(SessionID string) (User, bool) {

	var user User
	db, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	if SessionID != "" {
		err = base.DB.QueryRow("SELECT userID FROM Session WHERE sessionID = '" + SessionID + "'").Scan(&user.userID)
		if err == sql.ErrNoRows {
			return user, false
		} else {
			return user, true
		}
	} else {
		return user, false
	}
}

// func AreYouLogged(w http.ResponseWriter, r *http.Request) bool {
// 	c, err := r.Cookie("session")
// 	fmt.Println("here")
// 	if err != nil {
// 		return false
// 	}
// 	fmt.Println(c.Value)
// 	sess, ok := SessionDB[c.Value]
// 	fmt.Println(sess)
// 	fmt.Println(ok)
// 	if ok {
// 		//sess.timeOfLog = time.Now()
// 		SessionDB[c.Value] = sess
// 	}
// 	_, ok = UserDB[sess.SessionID]
// 	//c.MaxAge = 30
// 	http.SetCookie(w, c)
// 	return ok
// }

// more general functions.

// date as a global variable?
var date = time.Now().Format("2004.04.20 04:20:00")

// function to call into others to update the database info
func (base *Base) Update(table, set, to, where, id string) error {
	update := "UPDATE " + table + " SET " + set + " = '" + to + "' WHERE " + where + " = '" + id + "'"
	basedata, _ := base.DB.Prepare(update)
	_, err := basedata.Exec()
	if err != nil {
		return err
	}
	// fmt.Print(stmt)
	return nil
}

// https://www.tutorialspoint.com/sqlite/sqlite_delete_query.htm
// delete data from the database
func (base *Base) Delete(table, where, value string) error {
	// NEED TO PUT IN SPACES SO IT EXECUTES CORRECTLY
	remove := "DELETE FROM " + table + " WHERE " + where
	basedata, err := base.DB.Prepare(remove + " = (?)")
	if err != nil {
		return err
	}
	_, err = basedata.Exec(value)
	if err != nil {
		return err
	}
	return nil
}

// registration function
func (base *Base) Register(username, email, passw string) (string, string, string, error) {
	// date := time.Now().Format("2004.04.20 04:20:00")
	// how to do this not using UUID? use the built in UUID package but
	// doesnt seem to be easier? Not sure how to track. Cookies package?

	userID := uuid.NewV4()
	// passw, _ = CreateHash(passw)
	basedata, err := base.DB.Prepare(`
		INSERT INTO User (userID, username, email, password) values (?, ?, ?, ?)
	`)
	if err != nil {
		return "", "", "", err
		// log.Fatal(err)
	}
	_, err = basedata.Exec(userID, username, email, passw)
	if err != nil {
		// log.Fatal(err)
		return "", "", "", err
	}
	base.Update("User", "sessionID", "", "userID", userID.String())
	return userID.String(), username, "", nil
}

// func to start the session
func (base *Base) StartSession(userID string) (string, error) {
	// date := time.Now().Format("2004.04.20 04:20:00")
	sessionID := uuid.NewV4()
	// use quotes around things as seem to get an error for some things
	// maybe just because of the other SQL install you tried?
	// a way to avoid doing this for all funcs?
	basedata, _ := base.DB.Prepare(`
		INSERT INTO Session (sessionID, userID) values (?, ?)
	`)
	_, err := basedata.Exec(sessionID, userID)
	if err != nil {
		// needs two values so one can be a blank string
		// log.Fatal(err)
		// why does this make it not work?
		// just because its log.Fatal
		return "", err
	}
	base.Update("User", "sessionID", sessionID.String(), "userID", userID)
	return sessionID.String(), nil
}

func (base *Base) IsSessionValid(sessionID string) bool {
	sessuuid := ""
	err := base.DB.QueryRow("SELECT * FROM Session WHERE sessionID = '" + sessionID + "'").Scan(&sessuuid)
	if err == sql.ErrNoRows {
		// fmt.Print(err)
		return false
	}
	// initiate a new variable to compare against the sessionID
	// var inputedSession string
	// for horizontal.Next() {
	// 	horizontal.Scan(&inputedSession)
	// }
	return true
}

// delete session info using the previous function and sessionID
// https://www.sqlitetutorial.net/sqlite-delete/
// https://stackoverflow.com/questions/68322484/how-to-delete-row-in-go-sqlite3
func (base *Base) DeleteSession(sessionID string) error {
	// little unsure about the multiple calls of sessionID
	// table = user, where = sessionID, rest?
	err := base.Update("User", "sessionID", "", "sessionID", sessionID)
	if err != nil {
		return err
	}
	// table = session, where = sessionID, value = sessionID
	err = base.Delete("Session", "sessionID", sessionID)
	if err != nil {
		return err
	}
	return nil
}

// checks the database for the info, not disimilar from the first thing I tried
func (base *Base) LoginUser(userName, passw string) (string, string, string, error) {
	// establish a variable that talks to the struct

	var users User
	// cant use "" for User will it be alright without?
	userRow, err := base.DB.Query("SELECT * FROM User WHERE username = '" + userName + "'")
	if err != nil {
		return "", "", "", err
	}
	// more variables to talk to the struct
	// then relate them to each part of the struct
	// var usID, sesID, usNm, eMa, CreatedDate, pass string
	for userRow.Next() {
		userRow.Scan(&users.userID, &users.username, &users.email, &users.password, &users.SessionID, &users.LoggedIn)
		// users = User{
		// 	// need to differentiate them why no error yesterday?
		// 	// can this be done differently? linehk repo avoids this
		// 	// because using different version of SQL
		// 	// no it has to use only one reciever
		// 	userID:      usID,
		// 	SessionID:   sesID,
		// 	username:    usNm,
		// 	email:       eMa,
		// 	createdDate: CreatedDate,
		// 	password:    pass,
		// }
	}
	// basedata, _ := base.DB.Prepare("INSERT INTO User (LoggedIn) value (?)")
	// _, err2 := basedata.Exec(true)
	// if err2 != nil {
	// 	return "", "", "", err2
	// }
	// fmt.Println(passw == users.password)
	// fmt.Println(passw)
	// fmt.Println(users.password)

	// if the entry doesnt match the database
	// return the error
	if users.username == "" {
		return "", "", "", errors.New("USER NOT FOUND")
	}
	// checks the entered password against the hashed one
	// 24/07 not matching passwords correctly
	// if !(CheckHash(passw, users.password)) {
	if passw != users.password {
		return "", "", "", errors.New("UNMATCHED PASSWORDS")
	}
	// if not a new session/blank remove it
	// 29/07 is this why you're autologged out when clicking on something?
	if users.SessionID != "" {
		base.DeleteSession(users.SessionID)
	}
	// then create a new session
	seshion, err := base.StartSession(users.userID)
	if err != nil {
		return "", "", "", err
	}
	// pass this new session into the struct
	users.SessionID = seshion
	return users.userID, users.username, users.SessionID, nil
}

// funcs to deal with posts

// similar to the other funcs now
func (base *Base) PostPost(userID, title, category, category2, body string) (string, error) {
	// date := time.Now().Format("2004.04.20 04:20:00")
	postID := uuid.NewV4()
	basedata, err := base.DB.Prepare(`
	INSERT INTO Post (postID, userID, title, category, category2, datePosted, body) values (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return "", err
	}
	_, err = basedata.Exec(postID, userID, title, category, category2, date, body)
	if err != nil {
		return "", err
	}
	return postID.String(), nil
}

// two funcs for reactions? Can I make it one? Better to have two to distinguish
// basically the same code anyway just one pointing to posts other comments
// do I need to reaction database tables?
func (base *Base) ReactToPost(postID, userID string, reacted int) (string, error) {
	reactionID := uuid.NewV4()
	basedata, _ := base.DB.Prepare(`
	INSERT INTO Reaction (reactionID, postID, userID, reacted) values (?, ?, ?, ?)
	`)
	_, err := basedata.Exec(reactionID, postID, userID, reacted)
	if err != nil {
		log.Fatal(err)
	}
	return reactionID.String(), nil
}

// same func as IsCommentReactionValid
func (base *Base) IsPostReactionValid(posID, usID string) (string, int) {
	var reaction Reaction
	// var ReactionID, PostID, UserID string
	// var reactions int
	// maybe shouldnt do "" quotes? will test
	// dont use `` either?
	horizontal, err := base.DB.Query("SELECT reactionID, postID, userID, react FROM Reaction WHERE postID = '" + posID + "' AND userID = '" + usID + "' AND commentID IS NULL")
	// handle the error
	if err != nil {
		fmt.Print(err)
		return "", 0
	}
	// scans
	for horizontal.Next() {
		horizontal.Scan(&reaction.reactID, &reaction.postID, &reaction.userID, &reaction.numOfReacts)
		// reaction = Reaction{
		// 	reactID:     ReactionID,
		// 	postID:      PostID,
		// 	userID:      UserID,
		// 	numOfReacts: reactions,
		// }
	}
	return reaction.reactID, reaction.numOfReacts
}

func (base *Base) UpdatePostReaction(posID, usID, upDown string) {
	reID, validty := base.IsPostReactionValid(posID, usID)
	c, _ := strconv.Atoi(upDown)
	if validty == 0 {
		base.ReactToPost(posID, usID, c)
	} else if validty == c {
		base.Delete("Reaction", "reactID", reID)
	} else {
		base.Update("Reaction", "numOfReacts", upDown, "reactID", reID)
	}
}

// last func to make, basically the backend to homepage
// not last func as needed the other indexes
// need to select from the database by row to capture all the info
func (base *Base) PostIndex(sortBy, usID string) []map[string]interface{} {
	// var post Post
	var posts []map[string]interface{}
	// var postRows *sql.Rows
	// var err error
	// if usID != "" {
	// 	postRows, err = base.DB.Query("SELECT * FROM Post WHERE userID = '" + usID + "'")
	// 	if err != nil {
	// 		fmt.Print(err)
	// 		return posts
	// 	}
	// }

	postRows, err := base.DB.Query("SELECT * FROM Post")
	if err != nil {
		fmt.Print(err)
		return posts
	}

	var posID, uID, Title, subForum, subForum2, dateCreated, content interface{}
	for postRows.Next() {
		err = postRows.Scan(&posID, &uID, &Title, &subForum, &subForum2, &dateCreated, &content)
		posts = append(posts, map[string]interface{}{
			"postID":     posID,
			"userID":     uID,
			"title":      Title,
			"category":   subForum,
			"category2":  subForum2,
			"datePosted": dateCreated,
			"body":       content,
			// way to cast in comments to the struct
			// can be similar to these
			// "numComments": len(base.CommentIndex(posID)),
			// comment index
			// "comments": base.CommentIndex(posID),
			// need a func to get the reactions
			// eg reaction index
			// "reactions": base.PostReactionIndex(posID),
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		// var usersName string
		// userRow, err := base.DB.Query("SELECT username FROM User WHERE userID = '" + uID + "'")
		// if err != nil {
		// 	fmt.Print(err)
		// 	return posts
		// }
		// for userRow.Next() {
		// 	userRow.Scan(&usersName)
		// }
		// post.userID = usersName
		// // posts = append([]Post{post}, posts...)
		// // new func to sort or build in here?
		// switch sortBy {
		// case "frontend":
		// 	// if strings.Contains(subForum, "front end") {
		// 	// 	posts = append([]Post{post}, posts...)
		// 	// }
		// case "backend":
		// 	// if strings.Contains(subForum, "back end") {
		// 	// 	posts = append([]Post{post}, posts...)
		// 	// }
		// default:
		// 	// posts = append([]Post{post}, posts...)

		// }

	}
	return posts
}

// need 2 reaction indexes one for posts the other for comments

func (base *Base) PostReactionIndex(posID string) Reaction {
	var reaction Reaction
	// only two kinds of reaction so don't need the array here
	reactRows, err := base.DB.Query("SELECT reactID, postID, userID, numOfReacts FROM Reaction WHERE postID = '" + posID + "' AND commentID IS NULL")
	if err != nil {
		fmt.Print(err)
		return reaction
	}
	var rID, poID, usID string
	var reactsNum, upvote, downvote int
	for reactRows.Next() {
		reactRows.Scan(&rID, &poID, &usID, &reactsNum)
		reaction = Reaction{
			reactID:     rID,
			postID:      poID,
			userID:      usID,
			numOfReacts: reactsNum,
		}
		// able to streamline?
		if reactsNum == 1 {
			upvote++
		}
		if reactsNum == -1 {
			downvote++
		}
		reaction.upVotes = upvote
		reaction.downVotes = downvote
	}
	return reaction
}

// funcs to deal with Comments

func (base *Base) CommentComment(userID, postID, body string) (string, error) {
	// date := time.Now().Format("2004.04.20 04:20:00")
	commentID := uuid.NewV4()
	basedata, err := base.DB.Prepare(`
	INSERT INTO Comment (commentID, userID, postID, createdDate, body) values (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return "", err
	}
	_, err = basedata.Exec(commentID, userID, postID, date, body)
	if err != nil {
		return "", err
	}
	return commentID.String(), nil
}

// two funcs for reactions? Can I make it one? Better to have two to distinguish
// basically the same code anyway just one pointing to posts other comments
// do I need to reaction database tables?
func (base *Base) ReactToComment(postID, commentID, userID string, reacted int) (string, string, error) {
	reactionID := uuid.NewV4()
	basedata, _ := base.DB.Prepare(`
	INSERT INTO Reaction (reactionID, postID, commentID, userID, reacted) values (?, ?, ?, ?, ?)
	`)
	_, err := basedata.Exec(reactionID, postID, commentID, userID, reacted)
	if err != nil {
		log.Fatal(err)
	}
	return "", reactionID.String(), nil
}

// same as isPostReactionValid
// func (base *Base) IsCommentReactionValid(posID, usID string) (string, int) {
// 	var reaction Reaction
// 	var ReactionID, PostID, UserID, CommentID string
// 	var reactions int
// 	// maybe shouldnt do "" quotes? will test
// 	// dont use `` either?
// 	horizontal, err := base.DB.Query("SELECT FROM * Reaction WHERE commentID = '" + posID + "' AND userID = '" + usID + "'")
// 	// handle the error
// 	if err != nil {
// 		fmt.Print(err)
// 		return "", 0
// 	}

// 	for horizontal.Next() {
// 		horizontal.Scan(&ReactionID, &PostID, &UserID, &CommentID, &reactions)
// 		reaction = Reaction{
// 			reactID:     ReactionID,
// 			postID:      PostID,
// 			userID:      UserID,
// 			commentID:   CommentID,
// 			numOfReacts: reactions,
// 		}
// 	}
// 	return ReactionID, reaction.numOfReacts
// }

// copied from update post reaction

// func (base *Base) UpdateCommentReaction(posID, usID, comID, upDown string) {
// 	reID, validty := base.IsCommentReactionValid(posID, usID)
// 	c, _ := strconv.Atoi(upDown)
// 	if validty == 0 {
// 		base.ReactToComment(posID, usID, comID, c)
// 	} else if validty == c {
// 		base.Delete("Reaction", "reactID", reID)
// 	} else {
// 		base.Update("Reaction", "numOfReacts", upDown, "reactID", reID)
// 	}
// }

// create an index for the comments to return them in each post
// similar to creating an index for the posts in the homepage
func (base *Base) CommentIndex(poID string) []map[string]interface{} {
	// var comment Comment
	var comments []map[string]interface{}
	// var comRows *sql.Rows
	// var err error
	comRows, err := base.DB.Query("SELECT * FROM Comment WHERE postID = '" + poID + "'")
	if err != nil {
		fmt.Print(err)
		return comments
	}

	// one above is how to select from the postID
	// comRows, err := base.DB.Query("SELECT * FROM Comment")
	// if err != nil {
	// 	fmt.Print(err)
	// 	return comments
	// }

	var comID, posID, usID, dateCreated, content string
	for comRows.Next() {
		comRows.Scan(&comID, &posID, &usID, &dateCreated, &content)
		comments = append(comments, map[string]interface{}{
			"commentID":   comID,
			"postID":      posID,
			"userID":      usID,
			"createdDate": dateCreated,
			"body":        content,
			// another for reactions need to do the func
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		// var pID string
		// postRow, err := base.DB.Query("SELECT * FROM Post WHERE postID = '" + posID + "'")
		// if err != nil {
		// 	fmt.Print(err)
		// 	return comments
		// }
		// for postRow.Next() {
		// 	postRow.Scan(&pID)
		// }
		// comment.postID = pID
		//	comments = append([]Comment{comment}, comments...)

	}
	return comments
}

// copy pasted from postReactionINdex

// func (base *Base) CommentReactionIndex(cID string) Reaction {
// 	var reaction Reaction
// 	// only two kinds of reaction so don't need the array here
// 	reactRows, err := base.DB.Query("SELECT * FROM Reaction WHERE commentID = '" + cID + "'")
// 	if err != nil {
// 		fmt.Print(err)
// 		return reaction
// 	}
// 	var rID, poID, usID, comID string
// 	var reactsNum, upvote, downvote int
// 	for reactRows.Next() {
// 		reactRows.Scan(&rID, &poID, &usID, &reactsNum, &comID)
// 		reaction = Reaction{
// 			reactID:     rID,
// 			postID:      poID,
// 			userID:      usID,
// 			numOfReacts: reactsNum,
// 			commentID:   comID,
// 		}
// 		// able to streamline?
// 		if reactsNum == 1 {
// 			upvote++
// 		}
// 		if reactsNum == -1 {
// 			downvote++
// 		}
// 		reaction.upVotes = upvote
// 		reaction.downVotes = downvote
// 	}
// 	return reaction
// }

// sorting/filtering funcs here
// built in some to the PostIndex func as easier to pass to frontend control

// your post is easy similar to the other but just selecting from the logged in user

// func (base *Base) UsersOwnPosts(sortBy, usID string) []Post {
// 	var post Post
// 	var posts []Post
// 	postRows, err := base.DB.Query("SELECT * FROM Post WHERE userID = '" + usID + "'")
// 	if err != nil {
// 		fmt.Print(err)
// 		return posts
// 	}
// 	var pID, pTitle, subForum, dateCreated, content string
// 	for postRows.Next() {
// 		postRows.Scan(&pID, &pTitle, &subForum, &dateCreated, &content)
// 		post = Post{
// 			postID:      pID,
// 			createdDate: dateCreated,
// 			title:       pTitle,
// 			category:    subForum,
// 			body:        content,
// 		}
// 		posts = append([]Post{post}, posts...)
// 	}
// 	return posts
// }

// some global varibles to talk to some of the structs
var SessionDB = map[string]Session{}
var UserDB = map[string]User{}

// make the structs to call into, need to relate to the database
type Base struct {
	DB *sql.DB
}

// make them all strings so easier to read and post to

// struct to keep track of the users details
type User struct {
	userID    string
	username  string
	email     string
	password  string
	SessionID string
	LoggedIn  bool
}

type Session struct {
	SessionID string
	// identity  string
	// ipAddress string
	//timeOfLog time.Time
	//	UUID      string
	userID string
}

type Post struct {
	postID      string
	userID      string
	createdDate string
	title       string
	body        string
	category    string
	category2   string
	numComments int
	comments    []Comment
	reactions   Reaction
}

type Comment struct {
	commentID   string
	userID      string
	postID      string
	createdDate string
	body        string
	reactions   Reaction
}

type Reaction struct {
	reactID     string
	postID      string
	commentID   string
	userID      string
	numOfReacts int
	upVotes     int
	downVotes   int
}

// does it need to be this format or can I do it as presented?
// needs to be this format
// PRIMARY KEY ("userID"),PRIMARY KEY ("userID"),
//		FOREIGN KEY ("session id)

// https://www.sqlitetutorial.net/sqlite-foreign-key/

// initialise each table in the database
// func createUser(db *sql.DB) {
// 	basedata, _ := db.Prepare(`
// 	CREATE TABLE IF NOT EXISTS "User" (
// 		"userID"		TEXT NOT NULL UNIQUE PRIMARY KEY,
// 		"username"		VARCHAR(64) NOT NULL UNIQUE,
// 		"email"			NOT NULL UNIQUE,
// 		"password"		VARCHAR(255) NOT NULL,
// 		"sessionID"		TEXT,
// 		FOREIGN KEY (sessionID)
// 			REFERENCES "Session" ("sessionID")
// 		CHECK (length("username") >= 2 AND length("username") <= 24)
// 		CHECK (("email") LIKE '%_@_%._%')
// 		CHECK (length("password") >= 10)
// 	);
// 	 `)
// 	basedata.Exec()
// }

// //	"uuid"		NOT NULL,
// func createSession(db *sql.DB) {
// 	basedata, _ := db.Prepare(`
// 	CREATE TABLE IF NOT EXISTS "Session" (
// 		"sessionID"	TEXT UNIQUE PRIMARY KEY,
// 		"ipAddress"	TEXT NOT NULL,
// 		"timeOfLog"	TEXT NOT NULL,
// 		"userID"	TEXT NOT NULL UNIQUE,
// 		"identity"	TEXT NOT NULL,
// 		FOREIGN KEY (userID)
// 			REFERENCES "User" ("userID")
// 	);
// 	`)
// 	basedata.Exec()
// }

// func createPost(db *sql.DB) {
// 	basedata, _ := db.Prepare(`
// 	CREATE TABLE IF NOT EXISTS "Post" (
// 		"postID"	TEXT UNIQUE NOT NULL PRIMARY KEY,
// 		"userID"	TEXT NOT NULL,
// 		"title"     TEXT NOT NULL,
// 		"category"	TEXT NOT NULL,
// 		"datePosted" TEXT NOT NULL,
// 		"body"	TEXT NOT NULL,
// 		FOREIGN KEY ("userID")
// 			REFERENCES "User" ("userID")
// 	);
// 	`)
// 	basedata.Exec()
// }

// func createComment(db *sql.DB) {
// 	basedata, _ := db.Prepare(`
// 	CREATE TABLE IF NOT EXISTS "Comment" (
// 		"commentID" TEXT UNIQUE NOT NULL PRIMARY KEY,
// 		"postID"	TEXT NOT NULL,
// 		"userID"	TEXT NOT NULL,
// 		"dateCreated" TEXT NOT NULL,
// 		"body"	TEXT NOT NULL,
// 		FOREIGN KEY ("postID")
// 			REFERENCES "Post" ("postID")
// 		FOREIGN KEY ("userID")
// 			REFERENCES "User" ("userID")
// 	);
// 	`)
// 	basedata.Exec()
// }

// func createReaction(db *sql.DB) {
// 	basedata, _ := db.Prepare(`
// 	CREATE TABLE IF NOT EXISTS "Reaction" (
// 		"reactionID" TEXT NOT NULL PRIMARY KEY,
// 		"postID"	TEXT NOT NULL,
// 		"commentID" TEXT NOT NULL,
// 		"userID"	TEXT NOT NULL,
// 		"numOfReacts"		  INT,
// 		FOREIGN KEY ("postID")
// 			REFERENCES "Post" ("postID")
// 		FOREIGN KEY (commentID)
// 			REFERENCES "Comment" ("commentID")
// 		FOREIGN KEY ("userID")
// 			REFERENCES "User" ("userID")
// 	);
// 	`)
// 	basedata.Exec()
// }

func StartDatabase(db *sql.DB) *Base {
	// createUser(db)
	// createSession(db)
	// createPost(db)
	// createComment(db)
	// createReaction(db)
	return &Base{
		DB: db,
	}
}
