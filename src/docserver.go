package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type DocsPackage struct {
	// "foo"
	PackageId string

	// "/_packages/foo.tar.gz"
	DownloadKey string

	// "/docserver/_packages/foo-441d1f8e.tar.gz"
	StorageLocationWithVersion string
}

func makeServeLocation(packageId string) string {
	return "/docserver/" + packageId
}

func makeStorageLocationWithVersion(packageId string, md5Hash string) string {
	return "/docserver/_packages/" + packageId + "-" + md5Hash + ".tar.gz"
}

type test_struct struct {
	Test string
}

// /sync?token=..secret_token..
func httpSyncEndpoint(rw http.ResponseWriter, req *http.Request) {
	token := req.URL.Query().Get("token")

	requiredToken := "ef18d56168c3"

	if token == "" {
		http.Error(rw, "No token given", 400)
		return
	} else if token != requiredToken {
		http.Error(rw, "Token mismatch", 403)
		return
	}

	incomingRawJson, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	log.Println(string(incomingRawJson))
	var t test_struct
	err = json.Unmarshal(incomingRawJson, &t)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(rw, "ok thanks")
	// log.Println(t.Test)
}

func startStaticHttpServer() {
	http.HandleFunc("/sync", httpSyncEndpoint)
	http.Handle("/", http.FileServer(http.Dir("/docserver")))
	log.Fatal(http.ListenAndServe(":80", nil))
}

/*
	Serve /foo/index.html from /docserver/foo/index.html

	Foo directory is filled from foo.tar.gz
*/
func discoverPackagesFromS3(s3Session s3.S3) ([]DocsPackage, error) {
	// _packages/foo.tar.gz => foo
	packageRe := regexp.MustCompile(`_packages/([^\.]+)\.tar\.gz`)

	bucketName := "docs.function61.com"
	prefixFilter := "_packages/"

	packages := []DocsPackage{}

	err := s3Session.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: &bucketName,
		Prefix: &prefixFilter,
	}, func(listObjectsResult *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
		for _, s3Obj := range listObjectsResult.Contents {
			reMatch := packageRe.FindStringSubmatch(*s3Obj.Key)

			if reMatch != nil {
				packageId := reMatch[1]

				md5Hash := strings.Trim(*s3Obj.ETag, "\"")

				pack := DocsPackage{packageId, *s3Obj.Key, makeStorageLocationWithVersion(packageId, md5Hash)}

				packages = append(packages, pack)
			} else {
				// exclude the directory entry from "discarded" -warning
				if *s3Obj.Key != "_packages/" {
					log.Println("Discarded:", *s3Obj.Key)
				}
			}
		}

		return true
	})

	if err != nil {
		log.Println("failed to list objects", err)
		return nil, err
	}

	return packages, nil
}

func downloadPackage(pkg DocsPackage, s3Session s3.S3) {
	log.Println("Downloading ", pkg.DownloadKey)

	bucketName := "docs.function61.com"

	s3Response, err := s3Session.GetObject(&s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &pkg.DownloadKey,
	})
	if err != nil {
		panic(err)
	}

	defer s3Response.Body.Close()

	localFile, err := os.Create(pkg.StorageLocationWithVersion)
	if err != nil {
		panic(err)
	}
	defer localFile.Close()
	io.Copy(localFile, s3Response.Body)

	serveLocation := makeServeLocation(pkg.PackageId)

	log.Println("Downloaded. Extracting to", serveLocation)

	err = os.RemoveAll(serveLocation)
	if err != nil {
		panic(err)
	}

	err = os.Mkdir(serveLocation, os.FileMode(0775))
	if err != nil {
		panic(err)
	}

	extractTarGzCommand := exec.Command("tar", "-C", serveLocation, "-zxf", pkg.StorageLocationWithVersion)
	err = extractTarGzCommand.Run()
	if err != nil {
		panic(err)
	}

	log.Println("Extracted", pkg.PackageId)
}

func syncOnce() {
	sess, err := session.NewSession()
	if err != nil {
		log.Println("failed to create session,", err)
		return
	}

	s3Session := s3.New(sess, aws.NewConfig().WithRegion("us-east-1"))

	packages, err := discoverPackagesFromS3(*s3Session)
	if err != nil {
		panic(err)
	}

	for _, pkg := range packages {
		// TODO: handle other errors
		if _, err := os.Stat(pkg.StorageLocationWithVersion); os.IsNotExist(err) {
			log.Println("+ package does not exist", pkg.StorageLocationWithVersion)

			downloadPackage(pkg, *s3Session)
		} else {
			log.Println("  package exists", pkg.StorageLocationWithVersion)
		}
	}
}

func main() {
	go syncOnce()
	startStaticHttpServer()
}
