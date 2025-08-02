package models

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type Movie struct {
    Title string `bson:"title"`
    Year  string `bson:"year"`
    Link  string `bson:"link"`
    Image string `bson:"image"`
}


var movieCollection *mongo.Collection

func InitMongo(uri, dbName, collectionName string) error {
    timeOut := 10 * time.Second
    ctx, cancel := context.WithTimeout(context.Background(), timeOut)
    defer cancel()

    clientOpts := options.Client().ApplyURI(uri)

    client, err := mongo.Connect(ctx, clientOpts)

    if err != nil {
        return err
    }

    log.Println("MongoDB Connected")
    
    movieCollection = client.Database(dbName).Collection(collectionName)
    
    return nil
}

func InsertMovies(movies []Movie) error {
    if movieCollection == nil {
        return mongo.ErrClientDisconnected
    }

    timeOut := 15 * time.Second
    ctx, cancel := context.WithTimeout(context.Background(), timeOut)
    defer cancel()

    var docs []any
    for _, m := range movies {
        docs = append(docs, m)
    }

    _, err := movieCollection.InsertMany(ctx, docs)
    return err
}

func GetAllMovieLinks() ([]string, error) {
	if movieCollection == nil {
		return nil, mongo.ErrClientDisconnected
	}

        timeOut := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	cursor, err := movieCollection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"link": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var links []string
	for cursor.Next(ctx) {
		var m struct {
			Link string `bson:"link"`
		}
		if err := cursor.Decode(&m); err != nil {
			return nil, err
		}
		links = append(links, m.Link)
	}

	return links, nil
}

func ClearMoviesCollection() error {
    if movieCollection == nil {
        return mongo.ErrClientDisconnected
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()

    _, err := movieCollection.DeleteMany(ctx, bson.M{})
    return err
}


func HasMovies() (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()

    count, err := movieCollection.CountDocuments(ctx, bson.D{})
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

func EnsureTextIndex() error {
    if movieCollection == nil {
	return mongo.ErrClientDisconnected
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    mod := mongo.IndexModel{
	Keys: bson.D{{Key: "title", Value: "text"}}, // text index on "title"
    }

    _, err := movieCollection.Indexes().CreateOne(ctx, mod)
    return err
}

func SearchMoviesByTitle(query string) ([]Movie, error) {
    if movieCollection == nil {
        return nil, mongo.ErrClientDisconnected
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    filter := bson.M{
        "$text": bson.M{"$search": query},
    }

    cursor, err := movieCollection.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var results []Movie
    for cursor.Next(ctx) {
        var m Movie
        if err := cursor.Decode(&m); err != nil {
            return nil, err
        }
        results = append(results, m)
    }

    return results, nil
}

