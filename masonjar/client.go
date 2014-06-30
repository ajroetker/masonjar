package masonjar

import (
    "math/rand"
    "time"
    "appengine"
    "appengine/user"
    "net/http"
    "encoding/json"
)

type Board struct {
    Nertz []Card;
    Stream []Card;
    River [][]Card;
    Show []Card;
}

func init() {
    // Register our handlers with the http package.
    http.HandleFunc("/generate", generate)
}

func generate(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)

    data := new(Move)
    dec := json.NewDecoder(r.Body)
    dec.Decode(&data)

    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    enc.Encode(DealCards(NewShuffledDeck(u.String())))
}

func DealCards(deck <-chan Card) Board {
    nertz  := make([]Card, 13)
    stream := make([]Card, 35)
    river  := make([][]Card, 4)
    for pile := range river {
        river[pile] = make([]Card, 1)
        river[pile][0] = <-deck
    }
    show   := make([]Card, 0)
    for index := range nertz {
        nertz[index] = <-deck
    }
    for index := range stream {
        stream[index] = <-deck
    }
    board := Board{
        Nertz:  nertz,
        Stream: stream,
        River:  river,
        Show:   show,
    }
    return board
}

func NewShuffledDeck( name string ) <-chan Card {
    /* Generates a Deck as a process by creating an 'out' channel
    upon which the user of the Deck must input Cards */
    out := make(chan Card)
    go func() {

        // This is the Knuth Shuffle
        deck := make( []int , 52)
        for i := range deck {
            deck[i] = i
        }
        for i := 51; i >= 0; i-- {
            r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
            j := r.Intn(i+1)
            value := ( deck[j] % 13 ) + 1
            suit  := ( ( ( deck[j] - value ) + 1 ) / 13 ) + 1
            out <- Card{ value, suit, name, }
            if !( i == 0) { deck[j] = deck[i] };
        }
        close(out)
    }()
    return out
}
