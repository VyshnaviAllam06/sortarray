package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

// RequestPayload represents the structure of the incoming JSON payload
type RequestPayload struct {
	ToSort [][]int `json:"to_sort"`
}

// ResponsePayload represents the structure of the outgoing JSON response
type ResponsePayload struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNs       int64   `json:"time_ns"`
}

func main() {
	http.HandleFunc("/process-single", processSingle)
	http.HandleFunc("/process-concurrent", processConcurrent)
	http.ListenAndServe(":8000", nil)
}

func processSingle(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := make([][]int, len(requestPayload.ToSort))

	for i, subArray := range requestPayload.ToSort {
		sortedArrays[i] = make([]int, len(subArray))
		copy(sortedArrays[i], subArray)
		sort.Ints(sortedArrays[i])
	}

	timeTaken := time.Since(startTime).Nanoseconds()

	responsePayload := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNs:       timeTaken,
	}

	sendResponse(w, responsePayload)
}

func processConcurrent(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	var wg sync.WaitGroup
	sortedArrays := make([][]int, len(requestPayload.ToSort))
	mutex := &sync.Mutex{}

	for i, subArray := range requestPayload.ToSort {
		wg.Add(1)
		go func(i int, subArray []int) {
			defer wg.Done()
			sortedSubArray := make([]int, len(subArray))
			copy(sortedSubArray, subArray)
			sort.Ints(sortedSubArray)

			mutex.Lock()
			sortedArrays[i] = sortedSubArray
			mutex.Unlock()
		}(i, subArray)
	}

	wg.Wait()
	timeTaken := time.Since(startTime).Nanoseconds()

	responsePayload := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNs:       timeTaken,
	}

	sendResponse(w, responsePayload)
}

func sendResponse(w http.ResponseWriter, payload ResponsePayload) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
