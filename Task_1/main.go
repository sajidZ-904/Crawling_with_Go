package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func extractTextFromPDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var textBuilder strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		page := r.Page(i)
		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		textBuilder.WriteString(text)
		textBuilder.WriteString("\n")
	}

	return textBuilder.String(), nil
}

func parseQuestionBlock(block string) (bson.M, bool) {
	lines := strings.Split(block, "\n")
	var question string
	var options []string
	var answer, reference, concept string

	optionRegex := regexp.MustCompile(`\(\w+\)\s+.+`) // Matches options like "(a) Option text"
	answerRegex := regexp.MustCompile(`উত্তর:\s*(\([a-d]\))`)
	referenceRegex := regexp.MustCompile(`রেফারেন্স:|ররফাররন্স:`)
	conceptRegex := regexp.MustCompile(`কনসেপ্ট:|কনরেপ্ট:`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch {
		case answerRegex.MatchString(line):
			answerMatches := answerRegex.FindStringSubmatch(line)
			if len(answerMatches) > 1 {
				answer = answerMatches[1]
			}
		case referenceRegex.MatchString(line):
			reference = strings.TrimSpace(referenceRegex.Split(line, 2)[1])
		case conceptRegex.MatchString(line):
			concept = strings.TrimSpace(conceptRegex.Split(line, 2)[1])
		case optionRegex.MatchString(line):
			options = append(options, line)
		default:
			question += line + " "
		}
	}

	question = strings.TrimSpace(question)
	if question == "" && len(options) == 0 && answer == "" {
		fmt.Printf("Skipped Block: %s\n", block)
		return nil, false
	}

	// Ensure options have a minimum of 4 entries
	for len(options) < 4 {
		options = append(options, fmt.Sprintf("(%c) N/A", 'a'+len(options)))
	}

	return bson.M{
		"_id":        primitive.NewObjectID(),
		"question":   question,
		"options":    options,
		"answer":     answer,
		"references": reference,
		"concepts":   concept,
	}, true
}

func main() {
	filePath := "files/a.pdf"

	// Extract text from the PDF
	text, err := extractTextFromPDF(filePath)
	if err != nil {
		log.Fatalf("Failed to extract text from PDF: %v", err)
	}

	// Split text into question blocks using regex
	blockRegex := regexp.MustCompile(`\d{1,3}\.`)
	blocks := blockRegex.Split(text, -1)

	var bsonData []interface{}
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		parsedData, valid := parseQuestionBlock(block)
		if valid {
			bsonData = append(bsonData, parsedData)
		}
		if len(bsonData) == 100 { // Limit to 100 questions
			break
		}
	}

	// Insert data into MongoDB
	if len(bsonData) == 0 {
		log.Fatalf("No valid data to insert.")
	}

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("documents").Collection("questions")
	_, err = collection.InsertMany(context.Background(), bsonData)
	if err != nil {
		log.Fatalf("Failed to insert data: %v", err)
	}

	fmt.Println("Questions inserted successfully.")
}
