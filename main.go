package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Generate a unique filename using UUID
	uuidFilename := uuid.New().String() + filepath.Ext(handler.Filename)

	// Read the file content directly into a byte slice
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	// AWS S3 integration
	var (
		awsEndpoint = "http://localhost:9444/"
		//awsEndpoint    = "http://localhost:9444/newbucket"  <- Esto puede funcionar en AWS
		awsAccessKeyID = "AKIAIOSFODNN7EXAMPLE"
		awsSecretKey   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
		dummyRegion    = "us-east-1" // Placeholder region
		imageEndpoint  = "http://localhost:9444/ui/newbucket/"
	)

	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(awsEndpoint),
		Region:           aws.String(dummyRegion),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(awsAccessKeyID, awsSecretKey, ""),
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error creating AWS session:", err)
		return
	}

	svc := s3.New(sess)

	// Use just the UUID-generated filename as the object key
	objectKey := uuidFilename
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("newbucket"),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(fileBytes),
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error uploading to S3:", err)
		return
	}

	// Prepare data for the template
	data := struct {
		Filename string
		FileSize int64
		ImageURL string // Add ImageURL field
	}{
		Filename: uuidFilename,
		FileSize: handler.Size,
		ImageURL: imageEndpoint + strings.TrimPrefix(objectKey, "newbucket/"), // Construct the image URL
	}
	// Render the HTML template
	tmpl, err := template.ParseFiles("upload.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error parsing template:", err)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Error executing template:", err)
		return
	}
}

func setupRoutes() {
	http.HandleFunc("/upload", uploadFile)
	http.ListenAndServe(":8080", nil)
}

func main() {
	fmt.Println("Hello World")
	setupRoutes()
}
