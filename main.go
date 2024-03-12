package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type Video struct {
	VideoLibraryId       int      `json:"videoLibraryId"`
	Guid                 string   `json:"guid"`
	Title                string   `json:"title"`
	DateUploaded         string   `json:"dateUploaded"`
	Views                int      `json:"views"`
	IsPublic             bool     `json:"isPublic"`
	Length               int      `json:"length"`
	Status               int      `json:"status"`
	Framerate            float64  `json:"framerate"`
	Rotation             int      `json:"rotation"`
	Width                int      `json:"width"`
	Height               int      `json:"height"`
	AvailableResolutions string   `json:"availableResolutions"`
	ThumbnailCount       int      `json:"thumbnailCount"`
	EncodeProgress       int      `json:"encodeProgress"`
	StorageSize          int64    `json:"storageSize"`
	Captions             []string `json:"captions"`
	HasMP4Fallback       bool     `json:"hasMP4Fallback"`
	CollectionId         string   `json:"collectionId"`
	ThumbnailFileName    string   `json:"thumbnailFileName"`
	AverageWatchTime     int      `json:"averageWatchTime"`
	TotalWatchTime       int      `json:"totalWatchTime"`
	Category             string   `json:"category"`
	Chapters             []string `json:"chapters"` // Assuming chapters are strings; adjust if it's a complex type
	Moments              []string `json:"moments"`  // Same assumption as for chapters
	MetaTags             []string `json:"metaTags"`
	TranscodingMessages  []string `json:"transcodingMessages"`
}

type Payload struct {
	Title string `json:"title"`
}

func createVideo(title string) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	libraryId := os.Getenv("LIBRARYIDPROD")
	accessKey := os.Getenv("ACCESSKEYPROD")

	url := "https://video.bunnycdn.com/library/" + libraryId + "/videos"

	payload := Payload{
		Title: title,
	}

	// Marshal the payload into JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/*+json")
	req.Header.Add("AccessKey", accessKey)

	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != 200 {
		fmt.Println("Error occurred during API call. Status: ", res.StatusCode)
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))
	var video Video
	err = json.Unmarshal([]byte(body), &video)
	if err != nil {
		log.Fatal(err)
	}
	return video.Guid
}

func writeStringToFile(line string) {
	parts := strings.Split(line, ",")
	s := "./output/" + parts[0] + "_" + parts[1] + ".txt"
	file, err := os.Create(s)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close() // Ensure the file is closed after the function exits

	// Write the string to the file
	_, err = file.WriteString(line)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func uploadVideo(guid string, dir string, line string) {
	fmt.Println("Uploading video!")

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	libraryId := os.Getenv("LIBRARYIDPROD")
	accessKey := os.Getenv("ACCESSKEYPROD")

	println(libraryId, accessKey)

	parts := strings.Split(line, ",")
	username := parts[0]
	filename := parts[1]
	extention := parts[2]

	filePath := "./" + dir + "/" + username + "/" + filename + "." + extention
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read the file into a byte slice
	videoBytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	url := "https://video.bunnycdn.com/library/" + libraryId + "/videos/" + guid

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(videoBytes))

	req.Header.Add("accept", "application/json")
	req.Header.Add("AccessKey", accessKey)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	// body, _ := io.ReadAll(res.Body)
	// fmt.Println(string(body)) // {"success":true,"message":"OK","statusCode":200}
}

func loopThroughDirectory(dir string) {
	// read in a directory of of directories of videos
	// loop through each directory / video
	// Create a line entry with username,title
	// Write those entries to a file
	// Directory to loop through
	dirPath := "./" + dir

	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	// Loop through entries
	var allEntries [][]string
	nestedDirPath := ""
	for _, entry := range entries {
		if entry.IsDir() {

			nestedDirPath = dirPath + "/" + entry.Name()
			// Print directory name
			fmt.Println(nestedDirPath)

			videos, err := os.ReadDir(nestedDirPath)
			if err != nil {
				log.Fatal(err)
			}

			for _, video := range videos {
				if video.Name() == ".DS_Store" {
					continue
				}

				currentName := video.Name()
				fmt.Println(currentName)

				// Split the name to get the extension
				parts := strings.Split(currentName, ".")

				// Rename the video to a uuid
				id := uuid.New()
				newName := id.String() + "." + parts[1]

				// fullDirPath := nestedDirPath + "/" + entry.Name()
				err := os.Rename(nestedDirPath+"/"+currentName, nestedDirPath+"/"+newName)
				if err != nil {
					fmt.Println(err)
					log.Fatal("Error renaming file:", err)
				}

				final := []string{entry.Name(), id.String(), parts[1]}

				allEntries = append(allEntries, final)

			}
			fmt.Println("\n-------------------\n")

		}
	}

	// write to file
	// Open a file for writing
	file, err := os.Create("./to_upload.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Create a buffered writer from the file
	writer := bufio.NewWriter(file)

	// Iterate over the slice and write each element to the file
	for _, line := range allEntries {
		_, err := writer.WriteString(strings.Join(line, ",") + "\n") // Append a newline character after each line
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}

	// Make sure all buffered operations are applied to the file
	if err := writer.Flush(); err != nil {
		fmt.Println("Error flushing buffer to file:", err)
	}
}

func readCsvFile() []string {
	file, err := os.Open("to_upload.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return []string{}
	}
	defer file.Close() // Ensure the file is closed after finishing reading

	// Create a new buffered reader
	scanner := bufio.NewScanner(file)

	// Read the file line by line and save in array
	var lines []string
	for scanner.Scan() {
		line := scanner.Text() // Gets the current line
		lines = append(lines, line)
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	return lines
}

func work(wg *sync.WaitGroup, ch chan string, dir string) {
	for line := range ch {
		parts := strings.Split(line, ",")
		fmt.Println("Creating and uploading for ", parts[0], parts[1])
		guid := createVideo(parts[1])
		uploadVideo(guid, dir, line)
		newLine := line + "," + guid
		writeStringToFile(newLine)
		wg.Done()
	}
}

func main() {
	dir := "mainnet"
	// read in a directory of of directories of videos
	// loop through each directory / video
	// Create a line entry with username,title
	// Write those entries to a file
	loopThroughDirectory(dir)

	// Read the file and create a queue of videos to upload
	lines := readCsvFile()
	fmt.Println(lines)

	// Throw all those lines into a channel
	dataChannel := make(chan string)
	const maxNumWorkers = 5
	var wg sync.WaitGroup

	// Start workers
	for i := 1; i <= maxNumWorkers; i++ {
		go work(&wg, dataChannel, dir)
	}

	wg.Add(len(lines))
	// Send tasks to workers
	for _, line := range lines {
		dataChannel <- line
	}

	close(dataChannel)
	// Wait for all tasks to complete
	wg.Wait()
}
