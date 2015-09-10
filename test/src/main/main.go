package main

import (
	"flag"
	"fmt"
	"github.com/couchbaselabs/gocb"
	consul "github.com/hashicorp/consul/api"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

// these are populated by the flag variables
var (
	maxDocs     uint64
	concurrency int
)

func main() {
	doLoad := flag.Bool("load", false, "Load data into couchbase.")
	maxDocsVar := flag.Int("i", 1000, "Number of documents for load script. [default 1000]")
	concurrencyVar := flag.Int("c", 10, "Maximum number of goroutines to run. [default 10]")
	bucketVar := flag.String("b", "benchmark", "Name of bucket. [default benchmark]")
	// doViewTest := flag.Bool("view", false, "Load test via making view queries")
	// doQueryTest := flag.Bool("query", false, "Load test via making n1ql queries")
	flag.Parse()

	cluster := getCluster()
	bucket := getBucket(cluster, *bucketVar)
	if *doLoad {
		maxDocs = uint64(*maxDocsVar)
		concurrency = *concurrencyVar
		log.Printf("loading %v items w/ %v goroutines", maxDocs, concurrency)
		preLoadData(bucket)
		os.Exit(0)
	}

	// todo
	// if *doViewTest {
	// 	viewTest(bucket)
	// }
	// if *doQueryTest {
	// 	queryTest(bucket)
	// }
}

// query consul for the IP of the first CB node we find
// equivalent of `curl -s http://consul:8500/v1/catalog/service/couchbase`
func getClusterURL() string {
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
	url := fmt.Sprintf("couchbase://%s:8092", ip)
	log.Printf(url)
	return url
}

func getCluster() *gocb.Cluster {
	if cluster, err := gocb.Connect(getClusterURL()); err != nil {
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

func preLoadData(bucket *gocb.Bucket) {
	var wg sync.WaitGroup
	wg.Add(concurrency)

	reads := make(chan uint64)

	go func(reads chan uint64) {
		for i := uint64(0); i < maxDocs; i++ {
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

func loadDocuments(wg *sync.WaitGroup, reads chan uint64, bucket *gocb.Bucket) {
	for docId := range reads {
		if err := loadDoc(bucket, docId); err != nil {
			log.Println(err)
		}
	}
	wg.Done()
	return
}

func loadDoc(bucket *gocb.Bucket, docId uint64) error {
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
// utilities

// defer this function and pass the startTimer function in as its
// arguments and it will wrap the scope with a timer that writes
// out to log
func stopTimer(s string, startTime time.Time) {
	endTime := time.Now()
	log.Printf("time,%v,%v", s, endTime.Sub(startTime))
}

func startTimer(s string) (string, time.Time) {
	return s, time.Now()
}

func makeKey(docId uint64) string {
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

// ---------------------------------------------------
// debugging/testing

func checkDoc(bucket *gocb.Bucket, docId uint64) error {
	var key = fmt.Sprintf("%v", docId)
	var doc map[string]interface{}
	if cas, err := bucket.Get(key, &doc); err != nil {
		return err
	} else {
		log.Printf("got %v\n", doc["email"])
		doc["email"] = "newemail@joyent.com"
		if _, replaceErr := bucket.Replace(key, &doc, cas, 0); replaceErr != nil {
			return replaceErr
		} else {
			if _, secondGetErr := bucket.Get(key, &doc); secondGetErr == nil {
				log.Printf("got %v\n", doc["email"])
			} else {
				return secondGetErr
			}
		}
	}
	return nil
}
