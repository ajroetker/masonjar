package masonjar

import (
    "appengine"
    "appengine/user"
    "appengine/datastore"
    "net/http"
    "html/template"
    "fmt"
    "time"
)

func init() {
    // Register our handlers with the http package.
    http.HandleFunc("/", root)
    http.HandleFunc("/start/", start)
    http.HandleFunc("/leave/", leave)
    http.HandleFunc("/create/", create)
    http.HandleFunc("/remove/", remove)
    http.HandleFunc("/post", post)
}

// HTML template.
var tmpl = template.Must(template.ParseFiles("tmpl/index.html"))
var roottmpl = template.Must(template.ParseFiles("tmpl/root.html"))

// start is an HTTP handler that joins or creates a Game,
// creates a new Client, and writes the HTML response.
func root(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    url, err := user.LogoutURL(c, r.URL.String())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Create a list of games to display
    games, _ := getAllGames(c)
    htmlgames := ""
    for _, game := range games {
        var mark, status, quit, leave string
        if game.Status == 0 {
            mark = "ok"
            status = "success"
        } else {
            mark = "remove"
            status = "danger"
        }
        players, _ := game.GetPlayers(c)
        if len(players) == 0 {
            quit = "trash"
            leave = "remove"
        } else {
            quit = "remove-sign"
            leave = "leave"
        }
        htmlgames = fmt.Sprintf(`%v<tr class="%v"><td><a style="color:black;" href="/start/%v">%v</a></td><td>
          <span class="glyphicon glyphicon-%v"></span></td><td>%v</td>
          <td><a style="color:black;" href="/%v/%v"><span class="glyphicon glyphicon-%v"></span></a></td></tr>`,
          htmlgames, status, game.Id, game.Id, mark, len(players), leave, game.Id, quit)
    }

    // Render the HTML template
    data := struct{ Logout string; Games template.HTML }{ url, template.HTML(htmlgames) }
    err = roottmpl.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
}

// start is an HTTP handler that joins or creates a Game,
// creates a new Client, and writes the HTML response.
func create(w http.ResponseWriter, r *http.Request) {
    // Get the name from the request URL.
    name := r.URL.Path[8:]
    // If no valid name is provided, show an error.
    if !validName.MatchString(name) {
        http.Error(w, "Invalid tartan name", 404)
        return
    }
    c := appengine.NewContext(r)

    // Get or create the Game.
    _, err := getGame(c, name)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    // Wait for the db to update
    time.Sleep(100 * time.Millisecond)
    http.Redirect(w, r, "/", 301)
}

// start is an HTTP handler that joins or creates a Game,
// creates a new Client, and writes the HTML response.
func start(w http.ResponseWriter, r *http.Request) {
    // Get the name from the request URL.
    name := r.URL.Path[7:]
    // If no valid name is provided, show an error.
    if !validName.MatchString(name) {
        http.Error(w, "Invalid tartan name", 404)
        return
    }
    c := appengine.NewContext(r)

    // Get or create the Game.
    game, err := getGame(c, name)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Create a new Client, getting the channel token.
    token, err := game.AddPlayer(c, user.Current(c).String())
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    url, err := user.LogoutURL(c, r.URL.String())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Create a list of players to display
    players, _ := game.GetPlayers(c)
    // Send the current players the new list
    err = game.Send(c, Message{ Players : players } )
    if err != nil {
        http.Error(w, err.Error(), 500)
    }

    htmlplayers := ""
    for _, player := range players {
        htmlplayers = fmt.Sprintf(`%v<tr class="active">
        <td>%v</td>
        <td><a href="#"><span style="color:black;" class="glyphicon glyphicon-ok"></span></a></td>
        <td><a href="/leave/%v"><span style="color:black;" class="glyphicon glyphicon-remove"></span></a></td>
        </tr>`, htmlplayers, player.Name, game.Id)
    }

    // Render the HTML template
    data := struct{ Game, Token, Logout string; Players template.HTML }{ game.Id, token, url, template.HTML(htmlplayers) }
    err = tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

}

func remove(w http.ResponseWriter, r *http.Request) {
    // Get the name from the request URL.
    name := r.URL.Path[8:]
    // If no valid name is provided, show an error.
    if !validName.MatchString(name) {
        http.Error(w, "Invalid tartan name", 404)
        return
    }
    c := appengine.NewContext(r)

    // Remove the Game.
    game := &Game{ Status: 0, Id: name }
    err := game.remove(c)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    // Wait for the db to update
    time.Sleep(100 * time.Millisecond)
    http.Redirect(w, r, "/", 301)
}

func leave(w http.ResponseWriter, r *http.Request) {
    // Get the name from the request URL.
    name := r.URL.Path[7:]
    // If no valid name is provided, show an error.
    if !validName.MatchString(name) {
        http.Error(w, "Invalid tartan name", 404)
        return
    }
    c := appengine.NewContext(r)

    // Get the Game.
    game := &Game{ Id: name }

    err := datastore.Get(c, game.Key(c), game)
    if err != nil {
        http.Error(w, err.Error(), 500)
    }

    // Delete a Client.
    err = game.RemovePlayer(c, user.Current(c).String())
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    http.Redirect(w, r, "/", 301)
}

// post broadcasts a message to a specified Game.
func post(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)

    // Get the Game.
    game, err := getGame(c, r.FormValue("game"))
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Send the message to the clients in the game.
    msg := fmt.Sprintf( "%s: %s", u.String(), r.FormValue("msg") )
    players, err :=  game.GetPlayers(c)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    err = game.Send(c, Message{ Text : msg , Players : players } )
    if err != nil {
        http.Error(w, err.Error(), 500)
    }
}
