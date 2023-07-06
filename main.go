package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"rest/resthttp"
)

func main() {
	// Create a new instance of RestHttp
	client := resthttp.NewRestHttp("https://jsonplaceholder.typicode.com", "", "", true, false, 10*time.Second)

	// Perform a GET request
	container := "posts"
	resource := "1"
	response, err := client.GetRequest(container, resource, nil, "application/json", true)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the response body
	fmt.Println("Response:", string(response))

	// Perform a POST request
	container = "posts"
	resource = ""
	params := url.Values{}
	params.Set("title", "My Title")
	params.Set("body", "My Body")
	params.Set("userId", "1")
	response, err = client.PostRequest(container, resource, params, "application/json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the response body

	fmt.Println("Response:", string(response))

	// Perform a PUT request

	container = "posts"
	resource = "1"
	params = url.Values{}
	params.Set("title", "Updated Title")
	params.Set("body", "Updated Body")
	params.Set("userId", "1")
	response, err = client.PutRequest(container, resource, params, "application/json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the response body

	fmt.Println("Response:", string(response))

	// Perform a DELETE request

	container = "posts"
	resource = "1"
	response, err = client.DeleteRequest(container, resource, nil, "application/json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the response body

	fmt.Println("Response:", string(response))

	// Perform a GET request
	container = "photos"
	resource = "1"
	response, err = client.GetRequest(container, resource, nil, "application/json", true)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the response body

	fmt.Println("Response:", string(response))

	// Perform a file download

	container = "photos"
	resource = "1"
	savePath := "./downloaded.jpg"
	err = client.DownloadFile(container, resource, savePath, "", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("File downloaded successfully!")

	// Perform a file upload

	container = "photos"
	resource = ""
	filePath := "./upload.jpg"
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	params = url.Values{}
	params.Set("title", "Uploaded Photo")
	params.Set("albumId", "1")

	response, err = client.UploadFile(container, resource, params, "image/jpeg", file)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the response body
	fmt.Println("Response:", string(response))
}
