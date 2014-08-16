package masonjar

import (
    "net/http"
    "encoding/json"
    "appengine"
    "appengine/memcache"
)

type Move struct {
    To int
    Card Card
}

type CardGame struct {
    Lake []Card
    // Scores is a map from a players name to their score
    Scores map[string]int
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
    http.HandleFunc("/lake",  lake )
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

func (cg *CardGame) init(them []Player) {
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

    _, err := memcache.JSON.Get(c, "lake::nertz", &cardGame)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    enc.Encode(cardGame.Lake)
}

func begin(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    // Get or create the Game.
    board, err := getGame(c, "nertz")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Create a list of players to display
    players, err := board.GetPlayers(c)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    resp := checkReadinessOf(players)
    enc.Encode( map[string]bool{ "Valid": resp })

    if resp {
        var cardGame CardGame
        cardGame.init(players)
        err = memcache.JSON.Set(c, &memcache.Item{
            Key: "lake::nertz", Object: cardGame,
        })
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        // Send the current players the new list
        err = board.Send(c, Message{ Lake : cardGame.Lake } )
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
    }
}

func move(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    var cardGame CardGame

    _, err := memcache.JSON.Get(c, "lake::nertz", &cardGame)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    memcache.Delete(c, "lake::nertz")

    data := new(Move)
    dec := json.NewDecoder(r.Body)
    dec.Decode(&data)

    // Purge the now-invalid cache record (if it exists).

    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    resp := cardGame.attemptMove(*data)
    enc.Encode( map[string]bool{ "Valid": resp })

    if resp {
        // Get or create the Game.
        board, err := getGame(c, "nertz")
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        err = memcache.JSON.Set(c, &memcache.Item{
            Key: "lake::nertz", Object: cardGame,
        })
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        // Send the current players the new list
        err = board.Send(c, Message{ Lake : cardGame.Lake } )
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
    }
}

func end(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    var cardGame CardGame

    _, err := memcache.JSON.Get(c, "lake::nertz", &cardGame)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, "lake::nertz")

    // Get or create the Game.
    board, err := getGame(c, "nertz")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    //TODO render scoreboard
    //TODO Send the winners name
    // Send the current players the new list
    err = board.Send(c, Message{ Lake : cardGame.Lake } )
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
}
