package masonjar

import (
)

type CardGame struct {
    LakeChan chan []Card
    // Players is a map from a players name to their status
    Players map[string]int
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

func NewGame(them []Player) *CardGame {
    var players, scores map[string]int
    for _, player := range them {
        players[player.Name] = 0
        scores[player.Name] = 0
    }
    lakeChan := make(chan []Card, 1)
    lakeChan <- NewLake(len(them))
    return &CardGame{
        LakeChan : lakeChan,
        Players  : players,
        Scores   : scores,
    }
}

func (game *CardGame) attemptMove(card Card, pile int) bool {
    lake := <-game.LakeChan
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
}
