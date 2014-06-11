package masonjar

import (
    "appengine"
    "appengine/user"
    "appengine/datastore"
    "net/http"
    "html/template"
    "fmt"
    "strings"
)

type RowData struct{
    IconName, Activity, Game string
    NumPlayers int
}

type tmplData struct{
    PageName, Logout string
    TableData TableData
    Game, Token string
}

type TableData struct{
    Name string
    Headers []string
    Body []RowData
}

func init() {
    // Register our handlers with the http package.
    http.HandleFunc("/", root)
    http.HandleFunc("/start/", start)
    http.HandleFunc("/leave/", leave)
    http.HandleFunc("/ready/", ready)
    http.HandleFunc("/post", post)
}

// HTML template.
var tmpl = template.Must(template.ParseFiles("tmpl/index.html"))

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
    games, err := getAll(c)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    htmlgames := make([]RowData, len(games))
    for i, game := range games {
        players, _ := game.GetPlayers(c)
        if game.Status == 0 {
            htmlgames[i] = RowData{
            IconName : "ok",
            Activity : "success",
            NumPlayers: len(players),
            Game     : game.Id }
        } else {
            htmlgames[i] = RowData{
                IconName : "remove",
                Activity : "danger",
                NumPlayers: len(players),
                Game     : game.Id }
            }
        }

    // Render the HTML template
    data := tmplData{
        "Welcome to MasonJar!", url,
        TableData{
            "Current Games",
            []string{ "Name", "Ready", "#", "Remove", },
            htmlgames, },
        "","", }
    err = tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
}

func waiting(c appengine.Context, w http.ResponseWriter, r *http.Request, token string, game *Game) {

    url, err := user.LogoutURL(c, r.URL.String())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Create a list of players to display
    players, err := game.GetPlayers(c)
    if err != nil {
        http.Error(w, err.Error(), 500)
    }

    // Send the current players the new list
    err = game.Send(c, Message{ Players : players } )
    if err != nil {
        http.Error(w, err.Error(), 500)
    }

    htmlplayers := make([]RowData, len(players))
    for i, player := range players {
        if player.Status == 0 {
            htmlplayers[i] = RowData{
                IconName : "remove",
                Activity : "danger",
                // Set this to -1 for the template
                NumPlayers: -1,
                Game     : player.Name,
            }
        } else {
            htmlplayers[i] = RowData{
                IconName : "ok",
                Activity : "active",
                NumPlayers: -1,
                Game     : player.Name,
            }
        }
    }

    // Render the HTML template
    data := tmplData{
        game.Id, url,
        TableData{
             "Current Players",
             []string{ "Name", "Ready", },
             htmlplayers },
         game.Id, token, }
    err = tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

}

func ready(w http.ResponseWriter, r *http.Request) {
    // Get the name from the request URL.
    name := strings.Split(r.URL.Path, "/")[2]
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
    token, err := game.ReadyPlayer(c, user.Current(c).String())
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    waiting(c, w, r, token, game)
}

// start is an HTTP handler that joins or creates a Game,
// creates a new Client, and writes the HTML response.
func start(w http.ResponseWriter, r *http.Request) {
    // Get the name from the request URL.
    name := strings.Split(r.URL.Path, "/")[2]
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

    waiting(c, w, r, token, game)
}

func leave(w http.ResponseWriter, r *http.Request) {
    // Get the name from the request URL.
    name := strings.Split(r.URL.Path, "/")[2]
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

    // Create a list of players to display
    players, err := game.GetPlayers(c)
    if err != nil {
        http.Error(w, err.Error(), 500)
    }

    // Send the current players the new list
    err = game.Send(c, Message{ Players : players } )
    if err != nil {
        http.Error(w, err.Error(), 500)
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
