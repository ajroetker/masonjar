package masonjar

import (
    "appengine"
    "appengine/user"
    "net/http"
    "html/template"
    "fmt"
)

const clientIdLen = 40

func init() {
    // Register our handlers with the http package.
    http.HandleFunc("/", start)
    http.HandleFunc("/post", post)
}

// HTML template.
var tmpl = template.Must(template.ParseFiles("tmpl/index.html"))

// start is an HTTP handler that joins or creates a Room,
// creates a new Client, and writes the HTML response.
func start(w http.ResponseWriter, r *http.Request) {
    // Get the name from the request URL.
    name := r.URL.Path[1:]
    // If no valid name is provided, show an error.
    if !validName.MatchString(name) {
        http.Error(w, "Invalid tartan name", 404)
        return
    }
    c := appengine.NewContext(r)

    // Get or create the Room.
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
    htmlplayers := "<ul>"
    for _, player := range players {
        htmlplayers = fmt.Sprintf("%v<li>%v</li>", htmlplayers, player.Name)
    }
    htmlplayers = fmt.Sprintf("%v</ul>", htmlplayers)

    // Render the HTML template
    data := struct{ Room, Token, Logout string; Players template.HTML }{ game.Id, token, url, template.HTML(htmlplayers) }
    err = tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
}

// post broadcasts a message to a specified Room.
func post(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)

    // Get the room.
    //room, err := getRoom(c, r.FormValue("room"))
    room, err := getGame(c, r.FormValue("room"))
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Send the message to the clients in the room.
    err = room.Send(c, fmt.Sprintf( "%s: %s", u.String(), r.FormValue("msg")))
    if err != nil {
        http.Error(w, err.Error(), 500)
    }
}
