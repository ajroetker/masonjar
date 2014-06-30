package masonjar

import (
    "appengine"
    "appengine/user"
    "net/http"
    "html/template"
)

type tmplData struct{
    Logout, Game, Token string
}

func init() {
    // Register our handlers with the http package.
    http.HandleFunc("/", start)
    http.HandleFunc("/leave", MakeServePlayerStatusChanges(0) )
    http.HandleFunc("/ready", MakeServePlayerStatusChanges(1) )
    http.HandleFunc("/watch", MakeServePlayerStatusChanges(2) )
}

// HTML template.
var tmpl = template.Must(template.ParseFiles("tmpl/index.html"))


// start is an HTTP handler that joins or creates a Game,
// creates a new Client, and writes the HTML response.
func start(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    // Get or create the Game.
    game, err := getGame(c, "nertz")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Create a new Client, getting the channel token.
    token, err := game.GetPlayer(c, user.Current(c).String())
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
    players, err := game.GetPlayers(c)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Send the current players the new list
    err = game.Send(c, Message{ Players : players } )
    if err != nil {
        http.Error(w, err.Error(), 500)
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

func MakeServePlayerStatusChanges( status int ) func (w http.ResponseWriter, r *http.Request) {
    return func (w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        u := user.Current(c)

        // Get or create the Game.
        game, err := getGame(c, "nertz")
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        err = game.SetPlayerStatus(c, u.String(), status)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        players, err :=  game.GetPlayers(c)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        // Send the current players the new list
        err = game.Send(c, Message{ Players : players } )
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
    }
}
