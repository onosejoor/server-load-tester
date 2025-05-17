package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type ApiRes struct {
	Success bool
	Elapsed time.Duration
	Status  int
}

func main() {
	start := time.Now()

	url, numberOfUsers, err := getCmds()
	if err != nil {
		fmt.Println(err)
		return
	}

	var wg sync.WaitGroup
	c := make(chan ApiRes)

	for i := 0; i < numberOfUsers; i++ {
		wg.Add(1)
		go func() {
			testEndPoint(url, c, &wg)
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	successCount := 0
	total := 0

	for res := range c {
		total++
		if res.Success {
			successCount++
		}
		fmt.Printf("%+v\n", res)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✅ Program finished in %v\n", elapsed)
	fmt.Printf("🔁 Total requests: %d, ✅ Successes: %d, ❌ Failures: %d\n",
		total, successCount, total-successCount)
}

func testEndPoint(url string, c chan<- ApiRes, wg *sync.WaitGroup) error {

	defer wg.Done()

	start := time.Now()

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Get(url)
	if err != nil {
		log.Println("Error fetching URL:", err)
		c <- ApiRes{Success: false, Elapsed: time.Since(start)}
		return fmt.Errorf("error: %s", err)
	}
	defer res.Body.Close()

	var target any
	bodyBytes, _ := io.ReadAll(res.Body)
	status := res.StatusCode

	if err := json.Unmarshal(bodyBytes, &target); err != nil {
		log.Printf("JSON decode error: %v", err)
		c <- ApiRes{Success: false, Elapsed: time.Since(start), Status: status}
		return fmt.Errorf("error: %s", err)

	}

	c <- ApiRes{Success: true, Elapsed: time.Since(start), Status: status}
	return nil

}

func getCmds() (url string, numberOfUsers int, err error) {
	args := os.Args
	if len(args) < 2 {
		return "", 0, fmt.Errorf("usage: go run main.go <url> [number_of_users]")
	}

	url = args[1]
	count := "1"
	if len(args) >= 3 {
		count = args[2]
	}

	numberOfUsers, err = strconv.Atoi(count)
	if err != nil || numberOfUsers < 1 {
		return "", 0, fmt.Errorf("invalid number of users: %s", count)
	}

	return url, numberOfUsers, nil
}
