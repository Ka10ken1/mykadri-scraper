package models

import (
	"context"
	"log"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Show struct {
	Title        string `bson:"title"`
	TitleEnglish string `bson:"titleEnglish"`
	Year         string `bson:"year"`
	Link         string `bson:"link"`
	Image        string `bson:"image"`
	VideoURL     string `bson:"videoUrl"`
}

type ShowImage struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	Title        string             `bson:"title"`
	TitleEnglish string             `bson:"titleEnglish"`
	Image        string             `bson:"image" json:"image"`
}

var showCollection *mongo.Collection

func InitShowMongo(uri, dbName, collectionName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return err
	}

	log.Println("MongoDB Connected for shows")
	showCollection = client.Database(dbName).Collection(collectionName)
	return nil
}

func InsertShows(shows []Show) error {
	if showCollection == nil {
		return mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var docs []any
	for _, s := range shows {
		log.Printf("Inserting show: %+v\n", s)
		docs = append(docs, s)
	}

	_, err := showCollection.InsertMany(ctx, docs)
	return err
}

func GetAllShows() ([]Show, error) {
	if showCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := showCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var shows []Show
	for cursor.Next(ctx) {
		var s Show
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}
		shows = append(shows, s)
	}

	return shows, nil
}

func GetAllShowImages() ([]ShowImage, error) {
	if showCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetProjection(bson.M{
		"image":        1,
		"title":        1,
		"titleEnglish": 1,
	})

	cursor, err := showCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var images []ShowImage
	for cursor.Next(ctx) {
		var si ShowImage
		if err := cursor.Decode(&si); err != nil {
			return nil, err
		}
		images = append(images, si)
	}

	return images, nil
}

func GetShowByID(idStr string) (*Show, error) {
	if showCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var show Show
	err = showCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&show)
	if err != nil {
		return nil, err
	}

	return &show, nil
}

func GetAllShowLinks() ([]string, error) {
	if showCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := showCollection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"link": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var links []string
	for cursor.Next(ctx) {
		var s struct {
			Link string `bson:"link"`
		}
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}
		links = append(links, s.Link)
	}

	return links, nil
}

func ClearShowsCollection() error {
	if showCollection == nil {
		return mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := showCollection.DeleteMany(ctx, bson.M{})
	return err
}

func HasShows() (bool, error) {
	if showCollection == nil {
		return false, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := showCollection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func SearchShowsByTitle(query string) ([]Show, error) {
	if showCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	safeQuery := regexp.QuoteMeta(query)

	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": safeQuery, "$options": "i"}},
			{"titleEnglish": bson.M{"$regex": safeQuery, "$options": "i"}},
		},
	}

	cursor, err := showCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []Show
	for cursor.Next(ctx) {
		var s Show
		if err := cursor.Decode(&s); err != nil {
			return nil, err
		}
		results = append(results, s)
	}

	return results, nil
}

