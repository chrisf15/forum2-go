https://gosamples.dev/sqlite-intro/

use this to create all of it, basically a separate function for each
part of the database you need to create.

Might be long but theres some overlap it seems so should be a-OK.

homepage func last? Needs the most so can build the other pages first then
add the posts to it

Second link is an example of how to use it,
not in SQLite so slightly harder but gives you an idea of the number
of Funcs needed

so use switch r.Method to handle what you're allowed to do after logging in.

middleware functions to handle other parts

keep it as text using blank templates . fprint

5 different homepages. your posts. liked posts. 

sql rows store them in variables.  sql map golang

how to loop through sql rows


posts.html
{{range $post := .}}
   <h1>UserID: {{$post.UserID}}</h1>
// then repeat ones like this for all the pieces on information you need
{{end}}

have a text area to make an input from the user to make a new post.frontend

filter funcs back end

use get user func on the comment and post functions to block non users from commenting or posting

how to set the loggedIn bool to postivie or negative?

as log inn delete all previous sessions