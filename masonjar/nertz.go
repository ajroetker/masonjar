package masonjar

import (
    "net/http"
    "encoding/json"
    "appengine"
)

var game *CardGame = NewGame()

type Move struct {
    To int
    Card Card
}

type CardGame struct {
    LakeChan chan []Card
    // Scores is a map from a players name to their score
    Scores map[string]int
}

func init() {
    // Register our handlers with the http package.
    http.HandleFunc("/move",  HandleMove )
    http.HandleFunc("/begin", HandleBegin )
}

func NewLake(numPlayers int) []Card {
    lake := make([]Card, numPlayers * 4)
    for pile := range lake {
        lake[pile] = Card{ Value: 0, }
    }
    return lake
}

func NewGame() *CardGame {
    lakeChan := make(chan []Card, 1)
    scores   := make(map[string]int)
    return &CardGame{
        LakeChan : lakeChan,
        Scores   : scores,
    }
}

func (cg *CardGame) init(them []Player) {
    for _, player := range them {
        cg.Scores[player.Name] = 0
    }
    cg.LakeChan <- NewLake(len(them))
}

func (game *CardGame) attemptMove(c appengine.Context, card Card, pile int) bool {
    select {
    case lake := <-game.LakeChan:
        top := lake[pile]
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
            lake[pile] = card
            game.Scores[card.Owner]++
        }
        game.LakeChan<-lake
        return false
    default:
        return false
    }
}

func (game *CardGame) restart(them []Player) map[string]int {
    <-game.LakeChan
    finalScore := game.Scores
    var scores map[string]int
    for _, player := range them {
        scores[player.Name] = 0
    }
    game.LakeChan <- NewLake(len(them))
    return finalScore
}

func HandleBegin(w http.ResponseWriter, r *http.Request) {
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

    game.init(players)
    lake := <-game.LakeChan
    game.LakeChan<-lake
    c.Infof("%v", game)

    // Send the current players the new list
    err = board.Send(c, Message{ Lake : lake } )
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
}

func HandleMove(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    c.Infof("%v", game)

    data := new(Move)
    dec := json.NewDecoder(r.Body)
    dec.Decode(&data)

    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    enc.Encode( map[string]bool{ "Valid": game.attemptMove( c, data.Card, data.To ) })
}
