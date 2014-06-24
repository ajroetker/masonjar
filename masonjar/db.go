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
    Status int
}

type Card struct{
    // 1 := 'Ace', 2 := 'Two', ..., 13 := 'King'
    Value int
    // 1 := 'Spades', 2 := 'Hearts', 3 := 'Clubs', 4 := 'Diamonds'
    Suit  int
    // Useful for multiplayer games
    Owner string
}

// Game represents a card game.
type Game struct{
    Status int // 0 := not started, 1 := in progress, 2 := done
    Id    string // name of the game
}

func (g *Game) Key(c appengine.Context) *datastore.Key {
    return datastore.NewKey(c, "Game", g.Id, 0, nil)
}

func (p *Player) Key(c appengine.Context, game *Game) *datastore.Key {
    return datastore.NewKey(c, "Player", p.Name, 0, game.Key(c))
}

// AddClient puts a Client record to the datastore with the Room as its
// parent, creates a channel and returns the channel token.
func (g *Game) GetPlayer(c appengine.Context, id string) (string, error) {
    client := &Player{ Status: 0, Name: id }

    fn := func(c appengine.Context) error {
        err := datastore.Get(c, client.Key(c, g), client)
        if err == datastore.ErrNoSuchEntity {
            _, err = datastore.Put(c, client.Key(c, g), client)
        }
        return err
    }

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, g.Id)

    // datastore.RunInTransaction prevents a race condition
    // where two requests might try to make a room that both
    // to not exist. The failed transaction retries.
    err := datastore.RunInTransaction(c, fn, nil)
    if err != nil {
        return "", err
    }

    return channel.Create(c, id)
}

// AddClient puts a Client record to the datastore with the Room as its
// parent, creates a channel and returns the channel token.
func (g *Game) NotReadyPlayer(c appengine.Context, id string) (string, error) {
    client := &Player{ Name: id, Status: 0 }
    _, err := datastore.Put(c, client.Key(c, g), client)
    if err != nil {
        return "", err
    }

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, g.Id)

    return channel.Create(c, id)
}

func (g *Game) ReadyPlayer(c appengine.Context, id string) (string, error) {
    client := &Player{ Name: id, Status: 1 }
    _, err := datastore.Put(c, client.Key(c, g), client)
    if err != nil {
        return "", err
    }

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, g.Id)

    return channel.Create(c, id)
}

func (g *Game) WatcherPlayer(c appengine.Context, id string) (string, error) {
    client := &Player{ Name: id, Status: 2 }
    _, err := datastore.Put(c, client.Key(c, g), client)
    if err != nil {
        return "", err
    }

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, g.Id)

    return channel.Create(c, id)
}

//TODO run this in a transaction?
func (g *Game) RemovePlayer(c appengine.Context, id string) error {
    client := &Player{ Name: id }
    err := datastore.Delete(c, client.Key(c, g) )
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

// getGame fetches a Game by name from the datastore,
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

// readyGame fetches a Game by name from the datastore,
// and marks the game as in progress
func readyGame(c appengine.Context, id string) (*Game, error) {
    game   := &Game{ Status: 1, Id: id }
    _, err := datastore.Put(c, game.Key(c), game)

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, "games")
    return game, err
}
