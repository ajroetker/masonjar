package masonjar

import (
    "math/rand"
    "time"
    "appengine"
    "appengine/user"
    "net/http"
    "encoding/json"
    "appengine/memcache"
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
    http.HandleFunc("/reset", reset)
    http.HandleFunc("/save", save)
}

func reset(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)
    board := DealCards(NewShuffledDeck(u.String()))
    err := memcache.JSON.Set(c, &memcache.Item{
        Key: u.String(), Object: board,
    })
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    enc.Encode(board)
}

func generate(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)

    var board Board

    _, err := memcache.JSON.Get(c, u.String(), &board)
    if err != nil && err != memcache.ErrCacheMiss {
        http.Error(w, err.Error(), 500)
        return
    }
    if err == memcache.ErrCacheMiss {
        board = DealCards(NewShuffledDeck(u.String()))
        err := memcache.JSON.Set(c, &memcache.Item{
            Key: u.String(), Object: board,
        })
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
    }
    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    enc.Encode(board)
}

func save(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)

    board := new(Board)
    dec := json.NewDecoder(r.Body)
    dec.Decode(&board)
    err := memcache.JSON.Set(c, &memcache.Item{
        Key: u.String(), Object: board,
    })
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
}

func DealCards(deck <-chan Card) Board {
    show := make([]Card, 0)
    river  := make([][]Card, 4)
    for pile := range river {
        river[pile] = make([]Card, 1)
        DealToArray(deck, river[pile])
    }
    nertz  := make([]Card, 13)
    DealToArray(deck, nertz)
    stream := make([]Card, 35)
    DealToArray(deck, stream)
    board := Board{
        Nertz: nertz,
        Stream: stream,
        River: river,
        Show: show,
    }
    return board
}

func DealToArray(deck <-chan Card, array []Card) {
    for index := range array {
        array[index] = <-deck
    }
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
