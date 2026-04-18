package main

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"time"

	SDK "interoperability-sdk-golang/interoperability_bridge_golang"
)

type Result struct {
	Value interface{}
	Error error
}

func FetchPages(startPage, endPage int) []Result {
	count := endPage - startPage + 1
	results := make([]Result, count)
	
	// Helper to create a *string for the bridge
	strPtr := func(s string) *string { return &s }

	type indexedResult struct {
		index int
		res   Result
	}
	resChan := make(chan indexedResult, count)

	for i := 0; i < count; i++ {
		pageNum := startPage + i
		go func(idx, page int) {
			time.Sleep(time.Duration(rand.Intn(201)+50) * time.Millisecond)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			fetchDone := make(chan Result, 1)
			go func() {
				// Convert page int to string then to *string
				pageStr := fmt.Sprintf("%d", page)
				
				params := SDK.FilterParams{
					Page: strPtr(pageStr), 
				}
				res, err := SDK.FetchInteroperability(params)
				fetchDone <- Result{Value: res, Error: err}
			}()

			select {
			case <-ctx.Done():
				resChan <- indexedResult{idx, Result{Error: fmt.Errorf("timeout exceeded")}}
			case r := <-fetchDone:
				resChan <- indexedResult{idx, r}
			}
		}(i, pageNum)
	}

	for i := 0; i < count; i++ {
		ir := <-resChan
		results[ir.index] = ir.res
	}
	return results
}

func main() {
	fmt.Println("--- Bhilani Interop SDK (Golang Concurrency) ---")
	results := FetchPages(1, 5)

	for i, res := range results {
		pageNum := i + 1
		if res.Error != nil {
			fmt.Printf("Page %d: Failed (%v)\n", pageNum, res.Error)
			continue
		}

		v := reflect.ValueOf(res.Value)
		for v.Kind() == reflect.Ptr { v = v.Elem() }

		dataField := v.FieldByName("Data")
		if dataField.IsValid() && dataField.Kind() == reflect.Slice {
			itemCount := dataField.Len()
			
			if itemCount > 0 {
				fmt.Printf("Page %d: Success\n", pageNum)
				// Iterate through the slice of items
				for j := 0; j < itemCount; j++ {
					item := dataField.Index(j)
					for item.Kind() == reflect.Ptr || item.Kind() == reflect.Interface {
						item = item.Elem()
					}
					
					// Extract and print the Title field
					titleField := item.FieldByName("Title")
					if titleField.IsValid() {
						fmt.Printf("  - Title: %v\n", titleField.Interface())
					}
				}
			} else {
				// Get TotalPages for the "No Data" message
				totalPages := 0
				paginationField := v.FieldByName("Pagination")
				if paginationField.IsValid() {
					for paginationField.Kind() == reflect.Ptr { paginationField = paginationField.Elem() }
					tpField := paginationField.FieldByName("TotalPages")
					if tpField.IsValid() {
						if tpField.Kind() >= reflect.Uint && tpField.Kind() <= reflect.Uint64 {
							totalPages = int(tpField.Uint())
						} else {
							totalPages = int(tpField.Int())
						}
					}
				}
				fmt.Printf("Page %d: Success (No Data - Server has %d pages)\n", pageNum, totalPages)
			}
		}
	}
}