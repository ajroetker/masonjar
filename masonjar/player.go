package masonjar

import (
    "appengine"
    "net/http"
)

func init() {
    // Register our handlers with the http package.
    http.HandleFunc("/_ah/channel/disconnected/", disconnected )
    http.HandleFunc("/_ah/channel/connected/", connected )
}

func connected(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    clientID := r.FormValue("from")

    gameId := "nertz"

    // Get or create the Game.
    game, err := getGame(c, gameId)
    if err != nil {
        c.Errorf( "getting %v: %v", gameId, err )
    }

    players :=  game.getPlayers(c)

    // Send the current players the new list
    err = game.Send(c, Message{ Players : players } )
    if err != nil {
        c.Errorf( "sending message to %v: %v", game.Id, err )
    }

    c.Infof( "%v connected", clientID )
}

func disconnected(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    clientID := r.FormValue("from")

    gameId := "nertz"

    // Get or create the Game.
    game, err := getGame(c, gameId)
    if err != nil {
        c.Errorf( "getting %v: %v", gameId, err )
    }

    game.RemovePlayer( c, clientID )
    players :=  game.getPlayers(c)

    if len(players) == 0 {
        err = game.purge( c )
        c.Errorf( "purging %v: %v", game.Id, err )
    }

    // Send the current players the new list
    err = game.Send(c, Message{ Players : players } )
    if err != nil {
        c.Errorf( "sending message to %v: %v", game.Id, err )
    }

    c.Infof( "%v disconnected", clientID )
}
