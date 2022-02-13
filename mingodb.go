package mingodb

import (
	"errors"
	"reflect"

	"github.com/fatih/structs"
	bolt "go.etcd.io/bbolt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Database represents a MingoDB database connection.
type Database struct {
	Path string

	db *bolt.DB
}

// Open creates a new database connection at the path specified.
// If the path does not exist, it will be created.
func Open(path string) (*Database, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &Database{Path: path, db: db}, nil
}

// Close closes the database connection and cleans up any resources.
// Will block until all pending operations have completed.
func (db *Database) Close() error {
	return db.db.Close()
}

// Collection returns a DB collection object with the
// specified name. If the collection does not exist,
// it will be created.
func (db *Database) Collection(name string) (*Collection, error) {
	// Is the collection name empty?
	if name == "" {
		return nil, ErrEmptyBucketName
	}

	// If not, create it.
	err := db.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})
	if err != nil {
		return nil, err
	}

	// Return the collection object.
	return &Collection{db: db, name: name}, nil
}

// CollectionMust returns a DB collection object with the
// specified name. If the collection does not exist,
// it will be created. Note: This function wraps Collection()
// and panics if an error is returned.
func (db *Database) CollectionMust(name string) *Collection {
	c, err := db.Collection(name)
	if err != nil {
		panic(err)
	}
	return c
}

// Collection represents a collection of MingoDB documents.
type Collection struct {
	db   *Database
	name string
}

// Name returns the name of the collection.
func (c *Collection) Name() string {
	return c.name
}

// Database returns a pointer to the collection's parent database.
func (c *Collection) Database() *Database {
	return c.db
}

// Drop deletes the collection.
func (c *Collection) Drop() error {
	return c.db.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(c.name))
	})
}

// InsertOne inserts a single document into the collection.
// Returns the _id of the inserted document (if generated by the
// DB, will be of type primitive.ObjectID).
//
// Expects doc to be either a struct or a map[string]interface{}.
// Note that if doc is a struct, only expored fields will be stored.
func (c *Collection) InsertOne(doc interface{}) (InsertID, error) {
	// Validate the document. Is it a struct or a map?
	t := reflect.TypeOf(doc)
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Map {
		return nil, ErrInvalidType
	}

	// If it's a struct, convert it to a map.
	var m map[string]interface{}
	if t.Kind() == reflect.Struct {
		m = structs.Map(doc)
	}
	if t.Kind() == reflect.Map {
		var ok bool

		// Can the map be converted to a map[string]interface{}?
		m, ok = doc.(map[string]interface{})
		if !ok {
			return nil, ErrInvalidType
		}
	}

	// Check if doc has an _id field.
	// If not, generate one and add it to the doc.
	id, ok := m["_id"]
	if !ok {
		id = primitive.NewObjectID()
		m["_id"] = id
	}

	// Validate the id and marshal it into bytes.
	_, bid, err := bson.MarshalValue(id) // Also returns id's reflect type. Not currently used.
	if err != nil {
		return nil, err
	}

	// Marshal the document into bytes.
	bdoc, err := bson.Marshal(m)
	if err != nil {
		return nil, err
	}

	// Insert the document.
	err = c.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.name))
		return b.Put(
			bid,
			bdoc,
		)
	})
	if err != nil {
		return nil, err
	}

	// Return the _id of the inserted document.
	return id, nil
}

func (c *Collection) GetByID(id interface{}) (interface{}, error) {
	_, bid, err := bson.MarshalValue(id)
	if err != nil {
		return nil, err
	}

	var doc []byte
	err = c.db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.name))
		doc = b.Get(bid)
		if doc == nil {
			return errors.New("document not found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	err = bson.Unmarshal(doc, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// InsertMany inserts multiple documents into the collection.
// Returns an array of the inserted documents' _id values
// (If generated by the DB, will be of type primitive.ObjectID).
func (c *Collection) InsertMany(docs []interface{}) ([]InsertID, error) {
	return nil, nil
}

// Find returns (up to) multiple documents from the collection based on the
// filter provided.
func (c *Collection) Find(filter interface{}) (*MultiResult, error) {
	return nil, nil
}

// FindOne returns the first document (if any) that matches the filter.
func (c *Collection) FindOne(filter interface{}, result interface{}) (*SingleResult, error) {
	return nil, nil
}

// CountDocuments returns the number of documents that match the filter.
func (c *Collection) CountDocuments(filter interface{}) (int, error) {
	return 0, nil
}

// UpdateOne
func (c *Collection) UpdateOne(doc interface{}) (*UpdateResult, error) {
	return nil, nil
}

// UpdateMany
func (c *Collection) UpdateMany(docs []interface{}) (*UpdateResult, error) {
	return nil, nil
}

// DeleteOne deletes a single document into the collection based on the filter
func (c *Collection) DeleteOne(filter interface{}) (*DeleteResult, error) {
	return nil, nil
}

// DeleteMany inserts multiple documents into the collection.
func (c *Collection) DeleteMany(filter []interface{}) (*DeleteResult, error) {
	return nil, nil
}

// // Aggregate
// func (c *Collection) Aggregate(pipeline interface{}) (*MultiResult, error) {
// 	return nil, nil
// }
