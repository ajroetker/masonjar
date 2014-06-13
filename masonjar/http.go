package masonjar

import (
    "appengine"
    "appengine/user"
    "net/http"
    "html/template"
    "fmt"
)

type RowData struct{
    IconName, Activity, Name string
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
    http.HandleFunc("/", start)
    http.HandleFunc("/leave", leave)
    http.HandleFunc("/watch", watch)
    http.HandleFunc("/ready", ready)
    http.HandleFunc("/post", post)
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
        switch player.Status {
        case 0:
            htmlplayers[i] = RowData{
                IconName : "remove",
                Activity : "danger",
                // Set this to -1 for the template
                Name     : player.Name,
            }
        case 1:
            htmlplayers[i] = RowData{
                IconName : "ok",
                Activity : "success",
                Name     : player.Name,
            }
        default:
            htmlplayers[i] = RowData{
                IconName : "search",
                Activity : "info",
                Name     : player.Name,
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

func watch(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    // Get or create the Game.
    game, err := getGame(c, "nertz")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    _, err = game.WatcherPlayer(c, user.Current(c).String())
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
    }
}

func ready(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    // Get or create the Game.
    game, err := getGame(c, "nertz")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    _, err = game.ReadyPlayer(c, user.Current(c).String())
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
    }
}

func leave(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    game, err := getGame(c, "nertz")
    if err != nil {
        http.Error(w, err.Error(), 500)
    }

    // Delete a Client.
    _, err = game.AddPlayer(c, user.Current(c).String())
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
    }
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
