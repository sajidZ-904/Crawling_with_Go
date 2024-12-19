package main

import (
    "fmt"
    "log"

    "github.com/nguyenthenguyen/docx"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "context"
)

// Define a struct to hold extracted data
type ExtractedData struct {
    Questions []string `bson:"questions"`
    Images    []string `bson:"images"`
    Equations []string `bson:"equations"`
}

// Function to extract data from a .docx file
func extractFromDocx(filePath string) (*ExtractedData, error) {
    r, err := docx.ReadDocxFile(filePath)
    if err != nil {
        return nil, err
    }
    defer r.Close()

    doc := r.Editable()
    text := doc.GetContent()
    
    // Dummy extraction logic
    questions := []string{"Question 1?", "Question 2?"}
    images := []string{"image1.png", "image2.png"}
    equations := []string{"E = mc^2", "a^2 + b^2 = c^2"}

    return &ExtractedData{
        Questions: questions,
        Images:    images,
        Equations: equations,
    }, nil
}

// MongoDB connection and data insertion
func insertDataToMongo(data *ExtractedData) error {
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        return err
    }
    defer client.Disconnect(context.TODO())

    collection := client.Database("documents").Collection("collections")
    _, err = collection.InsertOne(context.TODO(), data)
    return err
}

func main() {
    // Example usage
    data, err := extractFromDocx("word_files/c.docx")
    if err != nil {
        log.Fatalf("Failed to extract data: %v", err)
    }

    err = insertDataToMongo(data)
    if err != nil {
        log.Fatalf("Failed to insert data to MongoDB: %v", err)
    }

    fmt.Println("Data successfully extracted and stored.")
}