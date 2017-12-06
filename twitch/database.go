package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

// Database is the struct for holding the main database and the two buckets
type Database struct {
	db *bolt.DB
}

// Webhook stores the data of a single webhook
type Webhook struct {
	Channel string `json:"channel"` // discord channel id
	ID      string `json:"id"`      // webhook id
	Token   string `json:"token"`   // webhook token
}

var (
	DB *Database
)

// NewDB returns a new database
func NewDB() *Database {
	boltDB, err := bolt.Open(
		"twitch.db",
		// read+write
		0600,
		&bolt.Options{
			// if file cannot be opened in x seconds, throw an error
			// this will fail if another instance of the service
			// has a read lock on the database file
			Timeout: 1 * time.Second,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	DB = &Database{
		db: boltDB,
	}
	err = DB.init()
	if err != nil {
		log.Fatal(err)
	}

	return DB
}

// AddChannel adds a twitch channel to motitor and adds the channelID + webhook to be notified
func (d *Database) AddChannel(twitchName, channel string, hook *Webhook) (err error) {
	err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bt("twitch-channels"))
		// add the twitch channel name to the twitch bucket so we know to ask for updates for it
		return b.Put(bt(twitchName), bt(""))
	})
	if err != nil {
		return
	}

	err = d.db.Update(func(tx *bolt.Tx) error {
		// get/make the bucket that holds all the discord webhooks receiving updates for a twitch channel
		b := tx.Bucket(bt("discord-webhooks"))
		n, err := b.CreateBucketIfNotExists(bt(twitchName))
		if err != nil {
			return err
		}

		// put the webhook data in the bucket
		return n.Put(bt(channel), bt(hook.ID+":"+hook.Token))
	})
	if err != nil {
		return
	}

	var n []byte
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bt("discord-channels"))
		n = b.Get(bt(channel))
		return nil
	})

	raw := []byte{}
	if n != nil {
		names := map[string]string{}
		err = json.Unmarshal(n, &names)
		if err != nil {
			return
		}

		names[channel] = ""
		raw, err = json.Marshal(names)
		if err != nil {
			return
		}

		err = d.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bt("discord-channels"))
			return b.Put(bt(channel), raw)
		})

		return
	}

	raw, err = json.Marshal(map[string]string{
		twitchName: "",
	})
	err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bt("discord-channels"))
		return b.Put(bt(channel), raw)
	})

	return
}

// DeleteWebhook deletes a webhook from a twitch channel
func (d *Database) DeleteWebhook(twitchName, wID, cID string) (err error) {
	rawNames := []byte{}
	err = d.db.Update(func(tx *bolt.Tx) error {
		err = d.incrementKey(tx.Bucket(bt("twitch-channels")), bt(twitchName), -1)
		if err != nil {
			return err
		}

		b := tx.Bucket(bt("discord-webhooks"))
		b = b.Bucket(bt(twitchName))
		err = b.Delete(bt(wID))
		if err != nil {
			return err
		}

		b = tx.Bucket(bt("discord-channels"))
		rawNames = b.Get(bt(cID))
		return nil
	})
	if err != nil {
		return
	}

	if rawNames != nil {
		names := map[string]string{}
		err = json.Unmarshal(rawNames, &names)
		if err != nil {
			return
		}

		delete(names, twitchName)
		rawNames, err = json.Marshal(names)
		if err != nil {
			return
		}

		err = d.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bt("discord-channels"))
			return b.Put(bt(cID), rawNames)
		})
	}

	return
}

// incrementKey increments a key by a given amount
// this can only be called within a valid write transaction
func (d *Database) incrementKey(b *bolt.Bucket, key []byte, amt int) error {
	toInc := b.Get(key)
	if toInc == nil {
		if amt < 1 {
			return nil
		}
		return b.Put(key, bt(strconv.Itoa(amt)))
	}

	cur, err := strconv.Atoi(string(toInc))
	if err != nil {
		return err
	}

	cur += amt
	if cur == 0 {
		return b.Delete(key)
	}
	return b.Put(key, bt(strconv.Itoa(cur)))
}

// GetAllTwitchChannels returns all the twitch channels being tracked
func (d *Database) GetAllTwitchChannels() (channels []string, err error) {
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bt("twitch-channels"))
		b.ForEach(func(k, v []byte) error {
			channels = append(channels, string(k))
			return nil
		})
		return nil
	})

	return
}

// GetWebhooksByTwitchName returns a slice of all the webhooks for a twitch channel
func (d *Database) GetWebhooksByTwitchName(twitchName string) (hooks []*Webhook, err error) {
	err = d.db.View(func(tx *bolt.Tx) error {
		// get bucket containing all the webhooks for a twitch channel
		b := tx.Bucket(bt("discord-webhooks"))
		if h := b.Bucket(bt(twitchName)); h != nil {
			// iterate over all the keys stored in the bucket, and append
			// to a slice of webhooks to be returned
			err = h.ForEach(func(k, v []byte) error {
				// append webhook to slice
				webhook := bytes.Split(v, bt(":"))
				if len(webhook) != 2 {
					return errors.New("incorrect webhook value")
				}
				hooks = append(hooks, &Webhook{string(k), string(webhook[0]), string(webhook[1])})
				return nil
			})
		}

		return nil
	})

	return
}

// GetUserByID returns a twitch user by their id
// it tries to see if the user is cached and if
// not calls the twitch api and caches response
func (d *Database) GetUserByID(id string) (userData *UserData, err error) {
	var user []byte
	err = d.db.View(func(tx *bolt.Tx) error {
		// guaranteed to exist
		twitchBucket := tx.Bucket(bt("twitch-channels"))
		// not guaranteed
		if userBucket := twitchBucket.Bucket(bt("user-data")); userBucket != nil {
			copy(user, userBucket.Get(bt(id)))
		}
		return nil
	})
	if err != nil {
		return
	}

	if user == nil {
		userData, err = API.GetUserByID(id)
		if err != nil {
			return
		}

		var raw []byte
		raw, err = json.Marshal(userData)
		if err != nil {
			return
		}

		err = d.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bt("twitch-channels"))
			u, err := b.CreateBucketIfNotExists(bt("user-data"))
			if err != nil {
				return err
			}
			return u.Put(bt(id), raw)
		})

		return
	}

	err = json.Unmarshal(user, userData)
	return
}

// GetGameByID returns a twitch game by it's ID
// it tries to see if the game is cached and if
// not calls the twitch api and caches response
func (d *Database) GetGameByID(id string) (gameData *GameData, err error) {
	var game []byte
	err = d.db.View(func(tx *bolt.Tx) error {
		// guaranteed to exist
		twitchBucket := tx.Bucket(bt("twitch-channels"))
		// not guaranteed
		if gameBucket := twitchBucket.Bucket(bt("game-data")); gameBucket != nil {
			copy(game, gameBucket.Get(bt(id)))
		}
		return nil
	})
	if err != nil {
		return
	}

	if game == nil {
		gameData, err = API.GetGameByID(id)
		if err != nil {
			return
		}

		var raw []byte
		raw, err = json.Marshal(gameData)
		if err != nil {
			return
		}

		err = d.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bt("twitch-channels"))
			u, err := b.CreateBucketIfNotExists(bt("game-data"))
			if err != nil {
				return err
			}
			return u.Put(bt(id), raw)
		})

		return
	}

	err = json.Unmarshal(game, gameData)
	return
}

// GetTwitchNamesByChannel return all the tracked twitch names from a discord channel
func (d *Database) GetTwitchNamesByChannel(cID string) (names []string, err error) {
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bt("discord-channels"))
		if c := b.Get(bt(cID)); c != nil {
			nameMap := map[string]string{}
			err = json.Unmarshal(c, &nameMap)
			if err != nil {
				return err
			}

			for i := range nameMap {
				names = append(names, i)
			}
		}
		return nil
	})

	return
}

func (d *Database) webhook404(hook *Webhook) (err error) {
	err = d.db.Update(func(tx *bolt.Tx) error {
		var err error
		discord := tx.Bucket(bt("discord-channels"))
		raw := discord.Get(bt(hook.Channel))
		err = discord.Delete(bt(hook.Channel))
		if err != nil {
			return err
		}

		twitchChannels := map[string]string{}
		err = json.Unmarshal(raw, &twitchChannels)
		if err != nil {
			return err
		}

		twitch := tx.Bucket(bt("twitch-channels"))
		for i := range twitchChannels {
			err = d.incrementKey(twitch, bt(i), -1)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return
}

// Close closes the current databse
func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) init() error {
	return d.db.Update(func(tx *bolt.Tx) error {
		var err error
		_, err = tx.CreateBucketIfNotExists(bt("twitch-channels"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists(bt("discord-webhooks"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists(bt("discord-channels"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})
}

// bt is a shortcut to typing []byte("random string")
func bt(s string) []byte {
	return []byte(s)
}
