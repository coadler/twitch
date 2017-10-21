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

	// bucket for holding the list of twitch names
	// we need to check for updates on
	twitch *bolt.Bucket

	// bucket for holding the discord webhooks
	// for each twitch channel
	discord *bolt.Bucket
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
	// add the twitch channel name to the twitch bucket so we know to ask for updates for it
	err = d.twitch.Put([]byte(twitchName), []byte("0"))
	if err != nil {
		return
	}

	// get/make the bucket that holds all the channels receiving updates for a twitch channel
	b, err := d.discord.CreateBucketIfNotExists([]byte(twitchName))
	if err != nil {
		return
	}

	// marshal the webhook so it can be saved in the bucket
	hookMarshaled, err := json.Marshal(hook)
	if err != nil {
		return
	}

	// put the webhook data in the bucket
	err = b.Put([]byte(channelID), hookMarshaled)
	return
}

// GetTwitchChannels returns all the twitch channels being tracked
func (d *Database) GetTwitchChannels() (channels []string, err error) {
	err = d.twitch.ForEach(func(k, v []byte) error {
		channels = append(channels, string(k))
		return nil
	})
	return
}

// GetWebhooks returns a slice of all the webhooks for a twitch channel
func (d *Database) GetWebhooks(twitchName string) (hooks []*Webhook, err error) {
	// get bucket containing all the webhooks for a twitch channel
	b, err := d.discord.CreateBucketIfNotExists([]byte(twitchName))
	if err != nil {
		return
	}

	var w *Webhook
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

	return
}

// GetUserByID returns a twitch user by their id
// it tries to see if the user is cached and if
// not calls the twitch api and caches response
func (d *Database) GetUserByID(id string) (*UserData, error) {
	b, err := d.twitch.CreateBucketIfNotExists([]byte("user-data"))
	if err != nil {
		return nil, err
	}

	u := b.Get([]byte(id))
	if u == nil || len(u) < 1 {
		userdata, err := twitchapi.GetUserByID(id)
		if err != nil {
			return nil, err
		}

		marshaled, err := json.Marshal(userdata)
		if err != nil {
			return nil, err
		}
		b.Put([]byte(id), marshaled)

		return userdata, nil
	}

	var userdata *UserData
	err = json.Unmarshal(u, &userdata)
	if err != nil {
		return nil, err
	}

	return userdata, nil
}

// Close closes the current databse
func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) init() error {
	return d.db.Update(func(tx *bolt.Tx) error {
		var err error
		d.twitch, err = tx.CreateBucketIfNotExists([]byte("twitch-channels"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		d.discord, err = tx.CreateBucketIfNotExists([]byte("discord-channels"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}
