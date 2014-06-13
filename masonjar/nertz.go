package masonjar

import (
    "net/http"
    "encoding/json"
)

type CardGame struct {
    LakeChan chan []Card
    // Scores is a map from a players name to their score
    Scores map[string]int
}

func NewLake(numPlayers int) []Card {
    lake := make([]Card, numPlayers * 4)
    for pile := range lake {
        lake[pile] = Card{ Value: 0, }
    }
    return lake
}

func initGame(them []Player) *CardGame {
    var scores map[string]int
    for _, player := range them {
        scores[player.Name] = 0
    }
    lakeChan := make(chan []Card, 1)
    lakeChan <- NewLake(len(them))
    return &CardGame{
        LakeChan : lakeChan,
        Scores   : scores,
    }
}

func (game *CardGame) attemptMove(card Card, pile int) bool {
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

func get(w http.ResponseWriter, r *http.Request) {
}

type Move struct {
    To int
    Card Card
}

func move(w http.ResponseWriter, r *http.Request) {
    //TODO wrap this fn
    univ := new(CardGame)
    data := new(Move)
    dec := json.NewDecoder(r.Body)
    dec.Decode(&data)

    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    enc.Encode( map[string]bool{ "Valid": univ.attemptMove( data.Card, data.To ) })
}
