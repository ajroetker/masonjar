package masonjar

import (
    "appengine"
    "appengine/user"
    "net/http"
    "html/template"
    "strconv"
)

type tmplData struct{
    Logout, Game, Token string
}

func init() {
    // Register our handlers with the http package.
    http.HandleFunc("/", start)
    http.HandleFunc("/status", status )
}


// HTML template.
var tmpl = template.Must(template.ParseFiles("tmpl/index.html"))


// start is an HTTP handler that joins or creates a Game,
// creates a new Client, and writes the HTML response.
func start(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)
    playerId := u.String()

    // Get or create the Game.
    game, err := getGame(c, "nertz")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Create a new Client, getting the channel token.
    token, err := game.makePlayer(c, playerId)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    url, err := user.LogoutURL(c, r.URL.String())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Render the HTML template
    data := tmplData{ url, game.Id, token, }
    err = tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
}

func status(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)
    playerId := u.String()
    gameId := "nertz"

    // Get or create the Game.
    game, err := getGame(c, gameId)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Status
    // -------------
    // not ready : 0
    // ready     : 1
    // watch     : 2
    status, err := strconv.Atoi(r.FormValue("status"))
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    err = game.SetPlayerStatus(c, playerId, status)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    players :=  game.getPlayers(c)

    // Send the current players the new list
    err = game.Send(c, Message{ Players : players } )
    if err != nil {
        c.Errorf( "sending message to %v: %v", gameId, err )
    }
}
