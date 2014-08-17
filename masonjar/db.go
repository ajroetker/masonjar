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
func (g *Game) makePlayer(c appengine.Context, id string) (string, error) {
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

func (g *Game) SetPlayerStatus(c appengine.Context, id string, status int) error {
    client := &Player{ Name: id, Status: status }
    _, err := datastore.Put(c, client.Key(c, g), client)

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, g.Id)

    return err
}

func (g *Game) RemovePlayer(c appengine.Context, playerId string) error {
    player := &Player{ Name: playerId }
    err := datastore.Delete(c, player.Key(c, g) )

    // Purge the now-invalid cache record (if it exists).
    memcache.Delete(c, g.Id)
    // This deletes the players game.
    // We can add this functionality back later
    // where the user could log in again after disconnecting.
    memcache.Delete(c, player.Name)

    return err
}

func (g *Game) getPlayers(c appengine.Context) []Player {
    var clients []Player

    _, err := memcache.JSON.Get(c, g.Id, &clients)
    if err != nil && err != memcache.ErrCacheMiss {
        c.Errorf("loading players from cache: %v", err)
    }

    if err == memcache.ErrCacheMiss {
        q := datastore.NewQuery("Player").Ancestor(g.Key(c))
        _, err = q.GetAll(c, &clients)
        if err != nil {
            c.Errorf("loading players from datastore: %v", err)
        }
        err = memcache.JSON.Set(c, &memcache.Item{
            Key: g.Id, Object: clients,
        })
        if err != nil {
            c.Errorf("loading players into cache: %v", err)
        }
    }

    return clients
}

type Message struct{
    Text string
    Scoreboard map[string]int
    Players []Player
    Lake []Card
    Error string
}

func (g *Game) Send(c appengine.Context, message Message) error {
    var clients []Player

    _, err := memcache.JSON.Get(c, g.Id, &clients)
    if err != nil && err != memcache.ErrCacheMiss {
        return err
    }

    if err == memcache.ErrCacheMiss {
        q := datastore.NewQuery("Player").Ancestor(g.Key(c))
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
    // A game should have status 0 by default
    // to indicate it hasn't started
    game := &Game{ Status: 0, Id: id }

    fn := func(c appengine.Context) error {
        err := datastore.Get(c, game.Key(c), game)
        if err == datastore.ErrNoSuchEntity {
            _, err = datastore.Put(c, game.Key(c), game)
        }
        return err
    }

    // datastore.RunInTransaction prevents a race condition
    // where two requests might try to make a room that both
    // find to not exist. The failed transaction retries.
    return game, datastore.RunInTransaction(c, fn, nil)
}

func (g *Game) start(c appengine.Context) error {
    return g.postStatus(c, 1)
}

func (g *Game) stop(c appengine.Context) error {
    return g.postStatus(c, 0)
}

func (g *Game) postStatus(c appengine.Context, status int ) error {
    g.Status = status
    _, err := datastore.Put(c, g.Key(c), g)
    switch status {
    case 0:
        c.Infof("%v is waiting to begin", g.Id)
    case 1:
        c.Infof("%v has begun", g.Id)
    }
    return err
}
