package masonjar

import (
    "net/http"
    "encoding/json"
    "appengine"
    "appengine/user"
    "appengine/memcache"
    "fmt"
    "strconv"
)

type Move struct {
    To int
    Card Card
}

type CardGame struct {
    Id string
    Lake []Card
    // Scores is a map from a players name to their score
    Scores map[string]int
    Penalties map[string]int
}

func (game *Game) purge( c appengine.Context ) error {
    lakeId := fmt.Sprintf( "lake::%s", game.Id )
    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, lakeId )
    return game.stop( c )
}

func checkReadinessOf( players []Player ) bool {
    if len(players) == 1 {
        return players[0].Status != 0
    }
    return checkReadinessOf(players[0:len(players)/2]) && checkReadinessOf(players[len(players)/2:])
}

func init() {
    // Register our handlers with the http package.
    http.HandleFunc("/move",  move )
    http.HandleFunc("/begin", begin )
    http.HandleFunc("/end",   end )
    http.HandleFunc("/check", check )
    http.HandleFunc("/lake",  lake )
    http.HandleFunc("/score", score )
}

func score(w http.ResponseWriter, r *http.Request) {
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

    remains, err := strconv.Atoi(r.FormValue("remains"))
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    var cardGame CardGame
    lakeId := fmt.Sprintf("lake::%v", game.Id)
    _, err = memcache.JSON.Get(c, lakeId, &cardGame)
    if err != nil {
        c.Errorf("getting %v: %v", lakeId, err )
        purgeErr := game.purge( c )
        if purgeErr != nil {
            c.Errorf( "purgin %v: %v", game.Id, purgeErr )
        }
        http.Error(w, err.Error(), 500)
        return
    }
    cardGame.Penalties[ playerId ] = remains * 2
    err = memcache.JSON.Set(c, &memcache.Item{
        Key: cardGame.Id, Object: cardGame,
    })
    if err != nil {
        c.Errorf("setting %v: %v", lakeId, err )
        http.Error(w, err.Error(), 500)
        return
    }

    if ( len( cardGame.Penalties ) == len( cardGame.Scores ) ) {
        for player, _ := range cardGame.Scores {
            cardGame.Scores[player] -= cardGame.Penalties[player]
        }
        err = game.Send(c, Message{ Scoreboard: cardGame.Scores } )
        if err != nil {
            c.Errorf( "sending message to %v: %v", game.Id, err )
        }
        err = game.purge( c )
        if err != nil {
            c.Errorf( "purgin %v: %v", game.Id, err )
        }
    }

}

func NewLake(numPlayers int) []Card {
    lake := make([]Card, numPlayers * 4)
    for pile := range lake {
        lake[pile] = Card{ Value: 0, }
    }
    return lake
}

func countReadyPlayers(them []Player) int {
    var total int = 0
    for _, player := range them {
        if player.Status == 1 {
            total += 1
        }
    }
    return total
}

func (cg *CardGame) init(id string, them []Player) {
    cg.Id = id
    cg.Penalties = make(map[string]int)
    cg.Scores = make(map[string]int)
    for _, player := range them {
        if player.Status != 2 {
            cg.Scores[player.Name] = 0
        }
    }
    numberOfReadyPlayers := countReadyPlayers(them)
    cg.Lake = NewLake(numberOfReadyPlayers)
}

func (game *CardGame) attemptMove(move Move) bool {
    card := move.Card
    pile := move.To
    top := game.Lake[pile]
    var valid bool;
    switch {
    case top.Value == 0 && card.Value == 1:
        // If the pile is empty we can add an Ace
        valid = true
    case card.Value == top.Value + 1 && card.Suit == top.Suit:
        // Make sure the card has the right value and suit
        valid = true
    default:
        valid = false
    }
    if valid {
        game.Lake[pile] = card
        game.Scores[card.Owner]++
    }
    return valid
}

func lake(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    var cardGame CardGame
    lakeId := fmt.Sprintf( "lake::%v", "nertz" )

    _, err := memcache.JSON.Get(c, lakeId, &cardGame)
    if err != nil && err != memcache.ErrCacheMiss {
        http.Error(w, err.Error(), 500)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    enc.Encode(cardGame.Lake)
}

func check(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    gameId := "nertz"

    // Get or create the Game.
    game, err := getGame(c, gameId)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    enc.Encode( map[string]int{ "State": game.Status })

}

func begin(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    gameId := "nertz"

    // Get or create the Game.
    game, err := getGame(c, gameId)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    err = game.stop(c)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Create a list of players to display
    players := game.getPlayers(c)

    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    resp := checkReadinessOf(players)
    enc.Encode( map[string]bool{ "Valid": resp } )

    if resp {
        var cardGame CardGame
        lakeId := fmt.Sprintf( "lake::%v", game.Id )
        cardGame.init(lakeId, players)
        err = memcache.JSON.Set(c, &memcache.Item{
            Key: cardGame.Id, Object: cardGame,
        })
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        err = game.start( c )
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        // Send the current players the new list
        err = game.Send(c, Message{ Lake : cardGame.Lake } )
        if err != nil {
            c.Errorf( "sending message to %v: %v", game.Id, err )
        }
    }
}

func move(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    var cardGame CardGame
    gameId := "nertz"
    lakeId := fmt.Sprintf("lake::%v", gameId )

    _, err := memcache.JSON.Get(c, lakeId, &cardGame)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    memcache.Delete(c, lakeId)

    data := new(Move)
    dec := json.NewDecoder(r.Body)
    dec.Decode(&data)

    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    resp := cardGame.attemptMove(*data)
    enc.Encode( map[string]bool{ "Valid": resp })

    if resp {

        err = memcache.JSON.Set(c, &memcache.Item{
            Key: cardGame.Id, Object: cardGame,
        })
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        game, err := getGame(c, "nertz")
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        // Send the current players the new list
        err = game.Send(c, Message{ Lake : cardGame.Lake } )
        if err != nil {
            c.Errorf( "sending message to %v: %v", game.Id, err )
        }
    }
}

func end(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    var cardGame CardGame
    gameId := "nertz"
    lakeId := fmt.Sprintf("lake::%v", gameId )

    _, err := memcache.JSON.Get(c, lakeId, &cardGame)
    if err != nil && err != memcache.ErrCacheMiss {
        http.Error(w, err.Error(), 500)
        return
    }

    if err == memcache.ErrCacheMiss {
        c.Errorf("no game to end")
        return
    }

    // Get or create the Game.
    game, err := getGame(c, gameId)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    err = game.Send(c, Message{ Text : "masonJar.gameOver", } )
    if err != nil {
        c.Errorf( "sending message to %v: %v", game.Id, err )
    }

}
