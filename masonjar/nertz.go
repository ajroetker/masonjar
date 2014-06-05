package masonjar

import (
    "appengine"
    "appengine/channel"
    "appengine/datastore"
    "appengine/memcache"
)

// Rooms are stored in the datastore to be the parent entity of many Clients,
// keeping all the participants in a particular chat in the same entity group.

// Player is a participant in a card Game.
type Player struct{
    Name   string
}

type Card struct{
    // 1 := 'Ace', 2 := 'Two', ..., 13 := 'King'
    Value int
    // 1 := 'Spades', 2 := 'Hearts', 3 := 'Clubs', 4 := 'Diamonds'
    Suit  int
    // Useful for multiplayer games
    Owner string
}

type Lake struct{
    Piles [][]Card
}

// Game represents a card game.
type Game struct{
    Status int // 0 := not started, 1 := in progress, 2 := done
    Id    string // name of the game
}

func (g *Game) Key(c appengine.Context) *datastore.Key {
    return datastore.NewKey(c, "Game", g.Id, 0, nil)
}

// AddClient puts a Client record to the datastore with the Room as its
// parent, creates a channel and returns the channel token.
func (g *Game) AddPlayer(c appengine.Context, id string) (string, error) {
    key := datastore.NewKey(c, "Player", id, 0, g.Key(c))
    client := &Player{ Name: id }
    _, err := datastore.Put(c, key, client)
    if err != nil {
        return "", err
    }

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, g.Id)

    return channel.Create(c, id)
}

//TODO run this in a transaction?
// AddClient puts a Client record to the datastore with the Room as its
// parent, creates a channel and returns the channel token.
func (g *Game) RemovePlayer(c appengine.Context, id string) error {
    key := datastore.NewKey(c, "Player", id, 0, g.Key(c))
    err := datastore.Delete(c, key)
    if err != nil {
        return err
    }

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, g.Id)

    return err
}

func (g *Game) GetPlayers(c appengine.Context) ( []Player, error ) {
    var clients []Player

    _, err := memcache.JSON.Get(c, g.Id, &clients)
    if err != nil && err != memcache.ErrCacheMiss {
        return nil, err
    }

    if err == memcache.ErrCacheMiss {
        q := datastore.NewQuery("Player").Ancestor(g.Key(c))
        _, err = q.GetAll(c, &clients)
        if err != nil {
            return nil, err
        }
        err = memcache.JSON.Set(c, &memcache.Item{
            Key: g.Id, Object: clients,
        })
        if err != nil {
            return nil, err
        }
    }

    return clients, nil
}

type Message struct{
    Lake []Card
    Players []Player
    Error string
    Text string
}

func (g *Game) Send(c appengine.Context, message Message) error {
    var clients []Player

    _, err := memcache.JSON.Get(c, g.Id, &clients)
    if err != nil && err != memcache.ErrCacheMiss {
        return err
    }

    if err == memcache.ErrCacheMiss {
        q := datastore.NewQuery("Client").Ancestor(g.Key(c))
        _, err = q.GetAll(c, &clients)
        if err != nil {
            return err
        }
        err = memcache.JSON.Set(c, &memcache.Item{
            Key: g.Id, Object: clients,
        })
        if err != nil {
            return err
        }
    }

    for _, client := range clients {
        err = channel.SendJSON(c, client.Name, message)
        if err != nil {
            c.Errorf("sending %q: %v", message, err)
        }
    }

    return nil
}

func (game *Game) remove(c appengine.Context) error {

    // Purge the now-invalid cache record (if it exists).
    err := memcache.Delete(c, "games")
    if err != nil {
        return err
    }
    return datastore.Delete(c, game.Key(c))
}

// getRoom fetches a Room by name from the datastore,
// creating it if it doesn't exist already.
func getGame(c appengine.Context, id string) (*Game, error) {
    game := &Game{ Status: 0, Id: id }

    fn := func(c appengine.Context) error {
        err := datastore.Get(c, game.Key(c), game)
        if err == datastore.ErrNoSuchEntity {
            _, err = datastore.Put(c, game.Key(c), game)
        }
        return err
    }

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, "games")

    // datastore.RunInTransaction prevents a race condition
    // where two requests might try to make a room that both
    // to not exist. The failed transaction retries.
    return game, datastore.RunInTransaction(c, fn, nil)
}

func getAllGames(c appengine.Context) ([]Game, error) {
    var games []Game

    memcache.Delete(c, "games")
    _, err := memcache.JSON.Get(c, "games", &games)
    if err != nil && err != memcache.ErrCacheMiss {
        return nil, err
    }

    if err == memcache.ErrCacheMiss {
        q := datastore.NewQuery("Game")
        _, err = q.GetAll(c, &games)
        if err != nil {
            return nil, err
        }
        err = memcache.JSON.Set(c, &memcache.Item{
            Key: "games", Object: games,
        })
        if err != nil {
            return nil, err
        }
    }

    return games, nil
}
