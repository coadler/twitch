package twitch

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

// Database is the struct for holding the main database and the two buckets
type Database struct {
	db *bolt.DB
}

// Webhook stores the data of a single webhook
type Webhook struct {
	ID    string
	Token string
}

var (
	twitchapi *Twitch
)

// NewDB returns a new database
func NewDB(t *Twitch) *Database {
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
	db := &Database{
		db: boltDB,
	}
	err = db.init()
	if err != nil {
		log.Fatal(err)
	}
	twitchapi = t

	return db
}

// AddChannel adds a twitch channel to motitor and adds the channelID + webhook to be notified
func (d *Database) AddChannel(twitchName, channelID string, hook Webhook) (err error) {
	err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("twitch-channels"))
		// add the twitch channel name to the twitch bucket so we know to ask for updates for it
		return b.Put([]byte(twitchName), []byte(""))
	})
	if err != nil {
		return
	}

	// marshal the webhook so it can be saved in the bucket
	raw, err := json.Marshal(hook)
	if err != nil {
		return
	}

	err = d.db.Update(func(tx *bolt.Tx) error {
		// get/make the bucket that holds all the channels receiving updates for a twitch channel
		b, err := tx.CreateBucketIfNotExists([]byte("discord-channels"))
		if err != nil {
			return err
		}
		// put the webhook data in the bucket
		return b.Put([]byte(channelID), raw)
	})

	return
}

// GetTwitchChannels returns all the twitch channels being tracked
func (d *Database) GetTwitchChannels() (channels []string, err error) {
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("twitch-channels"))
		b.ForEach(func(k, v []byte) error {
			channels = append(channels, string(k))
			return nil
		})
		return nil
	})

	return
}

// GetWebhooks returns a slice of all the webhooks for a twitch channel
func (d *Database) GetWebhooks(twitchName string) (hooks []*Webhook, err error) {
	var w *Webhook
	var makeBucket bool

	err = d.db.View(func(tx *bolt.Tx) error {
		// get bucket containing all the webhooks for a twitch channel
		b := tx.Bucket([]byte("discord-channels"))
		if h := b.Bucket([]byte(twitchName)); h != nil {
			// iterate over all the keys stored in the bucket, and append
			// to a slice of webhooks to be returned
			err = b.ForEach(func(k, v []byte) error {
				// parse json byte slice -> webhook struct
				err = json.Unmarshal(v, &w)
				if err != nil {
					return err
				}
				hooks = append(hooks, w)
				return nil
			})
		} else {
			makeBucket = true
		}
		return nil
	})
	// return if the bucket wasn't nil
	if !makeBucket || err != nil {
		return
	}

	err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("discord-channels"))
		_, err = b.CreateBucketIfNotExists([]byte(twitchName))
		return err
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
		twitchBucket := tx.Bucket([]byte("twitch-channels"))
		// not guaranteed
		if userBucket := twitchBucket.Bucket([]byte("user-data")); userBucket != nil {
			copy(user, userBucket.Get([]byte(id)))
		}
		return nil
	})
	if err != nil {
		return
	}

	if user == nil {
		userData, err = twitchapi.GetUserByID(id)
		if err != nil {
			return
		}

		var raw []byte
		raw, err = json.Marshal(userData)
		if err != nil {
			return
		}

		err = d.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("twitch-channels"))
			u, err := b.CreateBucketIfNotExists([]byte("user-data"))
			if err != nil {
				return err
			}
			return u.Put([]byte(id), raw)
		})
		return
	}

	err = json.Unmarshal(user, userData)
	return
}

// Close closes the current databse
func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) init() error {
	return d.db.Update(func(tx *bolt.Tx) error {
		var err error
		_, err = tx.CreateBucketIfNotExists([]byte("twitch-channels"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("discord-channels"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}
