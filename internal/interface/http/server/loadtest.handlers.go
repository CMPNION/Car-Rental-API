package server

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// loadTestHandler simulates various types of system overload for testing purposes
// This endpoint is ONLY for demonstrating alerts during the SRE defense
func (s *Server) loadTestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	testType := r.URL.Query().Get("type")
	if testType == "" {
		testType = "cpu"
	}

	duration := r.URL.Query().Get("duration")
	if duration == "" {
		duration = "10"
	}

	workers := r.URL.Query().Get("workers")
	if workers == "" {
		workers = "4"
	}

	var err error
	var durSeconds int
	var workerCount int

	durSeconds, err = strconv.Atoi(duration)
	if err != nil {
		durSeconds = 10
	}

	workerCount, err = strconv.Atoi(workers)
	if err != nil {
		workerCount = 4
	}

	// Cap values to prevent actual system damage
	if durSeconds > 60 {
		durSeconds = 60
	}
	if workerCount > 20 {
		workerCount = 20
	}

	var result string
	switch testType {
	case "cpu":
		result = simulateCPULoad(durSeconds, workerCount)
	case "memory":
		result = simulateMemoryLoad(durSeconds)
	case "db":
		result = simulateDBLoad(durSeconds, workerCount)
	case "goroutines":
		result = simulateGoroutineLeak(durSeconds)
	case "all":
		result = simulateAllTypes(durSeconds, workerCount)
	default:
		result = simulateCPULoad(durSeconds, workerCount)
	}

	response := map[string]interface{}{
		"status":     "load_test_started",
		"type":       testType,
		"duration_s": durSeconds,
		"workers":    workerCount,
		"message":    result,
		"warning":    "This endpoint is for SRE demonstration only!",
	}

	json.NewEncoder(w).Encode(response)
}

// simulateCPULoad creates heavy CPU computations
func simulateCPULoad(durationSec, workers int) string {
	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for time.Since(startTime) < time.Duration(durationSec)*time.Second {
				// Prime number calculation - CPU intensive
				calculatePrimes(10000)
				// Math operations
				for j := 0; j < 1000; j++ {
					_ = math.Sqrt(float64(j*workerID+1)) * math.Sin(float64(j))
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
	}()

	return fmt.Sprintf("CPU load: %d workers running for %d seconds", workers, durationSec)
}

// simulateMemoryLoad allocates large amounts of memory
func simulateMemoryLoad(durationSec int) string {
	startTime := time.Now()
	var allocations [][]byte

	go func() {
		for time.Since(startTime) < time.Duration(durationSec)*time.Second {
			// Allocate 10MB chunks
			data := make([]byte, 10*1024*1024)
			for i := range data {
				data[i] = byte(i % 256)
			}
			allocations = append(allocations, data)

			// Keep only last 50 allocations to prevent OOM
			if len(allocations) > 50 {
				allocations = allocations[1:]
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	return fmt.Sprintf("Memory load: allocating ~%d MB over %d seconds", 10*durationSec*10, durationSec)
}

// simulateDBLoad creates heavy database query patterns
func simulateDBLoad(durationSec, workers int) string {
	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for time.Since(startTime) < time.Duration(durationSec)*time.Second {
				// Simulate heavy DB operations with in-memory work
				for j := 0; j < 100; j++ {
					_ = calculatePrimes(5000)
				}
				time.Sleep(50 * time.Millisecond)
			}
		}(i)
	}

	go func() {
		wg.Wait()
	}()

	return fmt.Sprintf("DB simulation: %d workers running heavy queries for %d seconds", workers, durationSec)
}

// simulateGoroutineLeak creates many goroutines to simulate goroutine leak
func simulateGoroutineLeak(durationSec int) string {
	startTime := time.Now()
	var wg sync.WaitGroup

	go func() {
		count := 0
		for time.Since(startTime) < time.Duration(durationSec)*time.Second {
			// Create goroutines that block
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					// Block for a while
					time.Sleep(time.Duration(durationSec) * time.Second)
				}(count)
				count++
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	// Return current goroutine count
	runtime.Gosched()
	return fmt.Sprintf("Goroutine leak simulation: creating many blocked goroutines for %d seconds", durationSec)
}

// simulateAllTypes runs all load types simultaneously
func simulateAllTypes(durationSec, workers int) string {
	simulateCPULoad(durationSec, workers)
	simulateMemoryLoad(durationSec)
	simulateDBLoad(durationSec, workers)
	simulateGoroutineLeak(durationSec)

	return fmt.Sprintf("ALL load types: CPU+Memory+DB+Goroutines for %d seconds", durationSec)
}

// calculatePrimes performs CPU-intensive prime number calculations
func calculatePrimes(limit int) int {
	count := 0
	for num := 2; num <= limit; num++ {
		isPrime := true
		for i := 2; i*i <= num; i++ {
			if num%i == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			count++
		}
	}
	return count
}
