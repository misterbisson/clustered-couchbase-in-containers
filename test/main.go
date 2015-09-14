package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/couchbase/gocb"
	consul "github.com/hashicorp/consul/api"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

// these are populated by the flag variables
var (
	maxDocs     int32
	concurrency int
	bucketName  string
	debug       bool
)

// for connecting to Couchbase REST API
// TODO: make these come from command line args
const (
	cbCredsUsername = "Administrator"
	cbCredsPassword = "password"
)

func main() {

	rand.Seed(time.Now().Unix())

	doLoad := flag.Bool("load", false, "Load data into couchbase.")
	maxDocsFlag := flag.Int("i", 1000, "Number of documents for load script. [default 1000]")
	concurrencyFlag := flag.Int("c", 10, "Maximum number of goroutines to run. [default 10]")
	bucketFlag := flag.String("b", "benchmark", "Name of bucket. [default benchmark]")
	doViewTest := flag.Bool("view", false, "Load test via making view queries")
	doN1QLTest := flag.Bool("n1ql", false, "Load test via making n1ql queries")
	debugFlag := flag.Bool("debug", false, "Write debug-level logs")

	// parse arguments and assign to configuration
	flag.Parse()
	maxDocs = int32(*maxDocsFlag)
	concurrency = *concurrencyFlag
	bucketName = *bucketFlag
	debug = *debugFlag

	log.Println("Connecting to cluster...")
	cluster := getCluster()
	bucket := getBucket(cluster, bucketName)

	if *doLoad {
		log.Printf("Loading %v items w/ %v goroutines", maxDocs, concurrency)
		preLoadData(bucket)
		os.Exit(0)
	}
	if *doViewTest {
		log.Printf("Running view queries test w/ %v goroutines...", concurrency)
		viewTest(bucket)
		os.Exit(0)
	}
	if *doN1QLTest {
		log.Printf("Running n1ql queries test w/ %v goroutines...", concurrency)
		n1qlTest(bucket)
		os.Exit(0)
	}
}

// ---------------------------------------------------
// cluster and bucket connections

// query consul for the IP of the first CB node we find
// equivalent of `curl -s http://consul:8500/v1/catalog/service/couchbase`
func getClusterIP() string {
	var ip string
	config := &consul.Config{Address: "consul:8500"}
	if client, err := consul.NewClient(config); err != nil {
		log.Fatal(err)
	} else {
		catalog := client.Catalog()
		if service, _, err := catalog.Service("couchbase", "", nil); err != nil {
			log.Fatal(err)
		} else {
			if len(service) > 0 {
				ip = service[0].ServiceAddress
			} else {
				log.Fatal("no IPs found for couchbase")
			}
		}
	}
	return ip
}

func getCluster() *gocb.Cluster {
	url := fmt.Sprintf("couchbase://%s:8092", getClusterIP())
	if cluster, err := gocb.Connect(url); err != nil {
		log.Fatal(err)
	} else {
		return cluster
	}
	return nil // stupid gofmt; we're just crashing here
}

func getBucket(cluster *gocb.Cluster, bucket string) *gocb.Bucket {
	if bucket, err := cluster.OpenBucket(bucket, ""); err != nil {
		log.Fatal(err)
	} else {
		return bucket
	}
	return nil // stupid gofmt; we're just crashing here
}

// ---------------------------------------------------
// loading test data

func preLoadData(bucket *gocb.Bucket) {

	var wg sync.WaitGroup
	wg.Add(concurrency)

	reads := make(chan int32)

	go func(reads chan int32) {
		for i := int32(0); i < maxDocs; i++ {
			reads <- i
		}
		close(reads)
	}(reads)

	defer stopTimer(startTimer(fmt.Sprintf("preload")))

	for i := 0; i < concurrency; i++ {
		go loadDocuments(&wg, reads, bucket)
	}
	wg.Wait()
}

func loadDocuments(wg *sync.WaitGroup, reads chan int32, bucket *gocb.Bucket) {
	for docId := range reads {
		if err := loadDoc(bucket, docId); err != nil {
			log.Println(err)
		}
	}
	wg.Done()
	return
}

func loadDoc(bucket *gocb.Bucket, docId int32) error {
	key := makeKey(docId)
	doc := makeDoc(key)
	if cas, err := timedInsert(bucket, key, doc, 0); err == nil {
	} else {
		// got a conflicting document; we're just going to overwrite
		// here so to uncomplicate having multiple runs of the benchmark
		if _, err := bucket.Replace(key, doc, cas, 0); err != nil {
			return err
		} else {
		}
	}
	return nil
}

func timedInsert(bucket *gocb.Bucket, key string, value interface{}, expiry uint32) (gocb.Cas, error) {
	defer stopTimer(startTimer(fmt.Sprintf("insert:%v", key)))
	return bucket.Insert(key, value, expiry)
}

// ---------------------------------------------------
// testing view queries

func viewTest(bucket *gocb.Bucket) {

	if err := createViewQuery(bucket); err == nil {
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func(*gocb.Bucket) {
				for {
					emailAddress := getRandomEmailAddress()
					getViewQuery(bucket, emailAddress)
				}
			}(bucket)
		}
		wg.Wait()
	} else {
		log.Fatal(err)
	}
}

// Hit the REST API to create a view that maps email to [id, name].
func createViewQuery(bucket *gocb.Bucket) error {
	url := fmt.Sprintf("http://%s:8092/%s/_design/viewByEmail", getClusterIP(), bucketName)
	var body = `{
  "views": {
    "byEmail": {
      "map": "function (doc, meta) {emit(doc.email, [meta.id, doc.name]);}"
    }
  }
}`

	resp, err := restApiCall("PUT", url, body, "application/json")
	debugPrintf("%v", resp)
	return err
}

// Get the document associated with the emailAddress by hitting
// the view query we created above
func getViewQuery(bucket *gocb.Bucket, emailAddress string) error {

	byEmailQuery := gocb.NewViewQuery("viewByEmail", "byEmail")
	byEmailQuery.Key(emailAddress)
	var row interface{}

	defer stopTimer(startTimer(fmt.Sprintf("viewQuery:%v", emailAddress)))
	if rows, err := bucket.ExecuteViewQuery(byEmailQuery); err != nil {
		return err
	} else {
		rows.One(&row)
		debugPrintf("%v", row)
		return rows.Close()
	}
}

// ---------------------------------------------------
// testing n1ql queries

func n1qlTest(bucket *gocb.Bucket) {

	if err := prepareN1QLIndex(bucket); err == nil {
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func(*gocb.Bucket) {
				for {
					emailAddress := getRandomEmailAddress()
					if err := getN1QLQuery(bucket, emailAddress); err != nil {
						debugPrintf("%v", err)
						os.Exit(-1)
					}
				}
			}(bucket)
		}
		wg.Wait()
	} else {
		log.Fatal(err)
	}
}

// Hit the REST API to prepare a N1QL index
func prepareN1QLIndex(bucket *gocb.Bucket) error {
	url := fmt.Sprintf("http://%s:8093/query/service", getClusterIP())
	var body = fmt.Sprintf("statement=create primary index on %v", bucketName)

	if resp, err := restApiFormPost(url, body); err == nil {
		debugPrintf("%v", resp)
		body = fmt.Sprintf("statement=prepare select * from %v where email=$1", bucketName)
		resp, err = restApiFormPost(url, body)
		debugPrintf("%v", resp)
		return err
	} else {
		return err
	}
}

type N1QLResult struct {
	Results []json.RawMessage `json:"results"`
}

// Get the document associated with the emailAddress by making a N1QL query
func getN1QLQuery(bucket *gocb.Bucket, emailAddress string) error {
	url := fmt.Sprintf("http://%s:8093/query/service", getClusterIP())
	var body = fmt.Sprintf("statement=select * from %v where email='%v'", bucketName, emailAddress)
	defer stopTimer(startTimer(fmt.Sprintf("n1qlquery:%v", emailAddress)))
	if resp, err := restApiFormPost(url, body); err != nil {
		return err
	} else {
		if debug {
			var result N1QLResult
			if err := json.Unmarshal([]byte(resp), &result); err != nil {
				log.Printf(resp)
				return err
			}
			if len(result.Results) > 0 {
				log.Printf("%v", string(result.Results[0]))
			}
		}
	}
	return nil
}

// ---------------------------------------------------
// utilities

// defer this function and pass the startTimer function in as its
// arguments and it will wrap the scope with a timer that writes
// out to log

func restApiFormPost(url, body string) (string, error) {
	return restApiCall("POST", url, body, "application/x-www-form-urlencoded")
}

// wrap the REST API calls to Couchbase
func restApiCall(method, url, body, contentType string) (string, error) {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	req.SetBasicAuth(cbCredsUsername, cbCredsPassword)
	req.Header.Set("Content-Type", contentType)

	client := &http.Client{}
	if resp, err := client.Do(req); err != nil {
		return "", err
	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return string(body), nil
	}
}

func stopTimer(s string, startTime time.Time) {
	endTime := time.Now()
	log.Printf("time,%v,%vms", s, int64(endTime.Sub(startTime)/time.Millisecond))
}

func startTimer(s string) (string, time.Time) {
	return s, time.Now()
}

func makeKey(docId int32) string {
	return fmt.Sprintf("%020d", docId)
}

func makeDoc(docId string) map[string]string {
	var doc = map[string]string{
		"email": fmt.Sprintf("%v@joyent.com", docId),
		"name":  getRandomString(20),
	}
	return doc
}

func getRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// generates a random email address within the maxDocs range, for use
// in testing queries
func getRandomEmailAddress() string {
	docId := rand.Int31n(maxDocs)
	return fmt.Sprintf("%020d@joyent.com", docId)
}

func debugPrintf(fmtString string, vals ...interface{}) {
	if debug {
		log.Printf(fmtString, vals...)
	}
}
