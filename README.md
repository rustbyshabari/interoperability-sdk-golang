Basic Usage

    package main
    
    import (
    	"fmt"
    
    	SDK "interoperability-sdk-golang/interoperability_bridge_golang"
    )
    
    func RunDemo() {
    	strPtr := func(s string) *string { return &s }
    
    	fmt.Println("Go SDK")
    
    	params := SDK.FilterParams{
    		Page: strPtr("1"),
    	}
    
    	response, err := SDK.FetchInteroperability(params)
    	if err != nil {
    		fmt.Printf("Error: %v\n", err)
    		return
    	}
    
    	fmt.Printf("%+v\n", response)
    }
    
    func main() {
    	RunDemo()
    }

Dynamic Usage

    package main
    
    import (
    	"fmt"
    
    	SDK "interoperability-sdk-golang/interoperability_bridge_golang"
    )
    
    func FetchPage(page int) (SDK.FilterResponse, error) {
    	strPtr := func(s string) *string { return &s }
    	
    	pageStr := fmt.Sprintf("%d", page)
    	params := SDK.FilterParams{
    		Page: strPtr(pageStr),
    	}
    	
    	return SDK.FetchInteroperability(params)
    }
    
    func main() {
    	fmt.Println("--- Bhilani Interop SDK ---")
    
    	for pageNum := 1; pageNum <= 5; pageNum++ {
    		response, err := FetchPage(pageNum)
    		if err != nil {
    			fmt.Printf("Page %d: Failed (Error: %v)\n", pageNum, err)
    			continue
    		}
    
    		totalPages := 0
    		if response.Pagination != nil {
    			totalPages = int(response.Pagination.TotalPages)
    		}
    
    		if len(response.Data) == 0 || pageNum > totalPages {
    			fmt.Printf("Page %d: Success (No Data - Server has %d pages)\n", pageNum, totalPages)
    		} else {
    			fmt.Printf("Page %d: Success\n", pageNum)
    			for _, item := range response.Data {
    				fmt.Printf("  - Title: %s\n", item.Title)
    			}
    		}
    	}
    }

Concurrent Usage

    package main
    
    import (
    	"context"
    	"fmt"
    	"math/rand"
    	"sync"
    	"time"
    
    	SDK "interoperability-sdk-golang/interoperability_bridge_golang"
    )
    
    type FetchResult struct {
    	PageNum int
    	Data    SDK.FilterResponse
    	Err     error
    }
    
    func FetchPages(start, end int) []FetchResult {
    	count := end - start + 1
    	results := make([]FetchResult, count)
    	var wg sync.WaitGroup
    	
    	strPtr := func(s string) *string { return &s }
    
    	for i := 0; i < count; i++ {
    		wg.Add(1)
    		go func(idx int) {
    			defer wg.Done()
    			pageNum := start + idx
    
    			time.Sleep(time.Duration(rand.Intn(201)+50) * time.Millisecond)
    
    			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    			defer cancel()
    
    			resChan := make(chan struct {
    				res SDK.FilterResponse
    				err error
    			}, 1)
    
    			go func() {
    				params := SDK.FilterParams{Page: strPtr(fmt.Sprintf("%d", pageNum))}
    				res, err := SDK.FetchInteroperability(params)
    				resChan <- struct {
    					res SDK.FilterResponse
    					err error
    				}{res, err}
    			}()
    
    			select {
    			case <-ctx.Done():
    				results[idx] = FetchResult{PageNum: pageNum, Err: fmt.Errorf("timeout exceeded")}
    			case r := <-resChan:
    				results[idx] = FetchResult{PageNum: pageNum, Data: r.res, Err: r.err}
    			}
    		}(i)
    	}
    
    	wg.Wait()
    	return results
    }
    
    func main() {
    	fmt.Println("--- Bhilani Interop SDK (Golang Concurrency) ---")
    
    	results := FetchPages(1, 5)
    
    	for _, res := range results {
    		if res.Err != nil {
    			fmt.Printf("Page %d: Failed (%v)\n", res.PageNum, res.Err)
    			continue
    		}
    
    		totalPages := 0
    		if res.Data.Pagination != nil {
    			totalPages = int(res.Data.Pagination.TotalPages)
    		}
    
    		if len(res.Data.Data) == 0 || res.PageNum > totalPages {
    			fmt.Printf("Page %d: Success (No Data - Server has %d pages)\n", res.PageNum, totalPages)
    		} else {
    			fmt.Printf("Page %d: Success\n", res.PageNum)
    			for _, item := range res.Data.Data {
    				fmt.Printf("  - Title: %s\n", item.Title)
    			}
    		}
    	}
    }

<img width="934" height="442" alt="Screenshot (207)" src="https://github.com/user-attachments/assets/5a54928b-ea9c-48b6-8fa0-f0b7825d7f5a" />
