package server_status

import (
	"context"
	"fmt"
	"github.com/veteran-software/yourddo-api/shared/types"
	"io"
	"net/http"
	"sync"
	"time"
)

type WorkerPool struct {
	workers    int
	maxRetries int
	client     *http.Client
}

func NewWorkerPool(workers, maxRetries int) *WorkerPool {
	return &WorkerPool{
		workers:    workers,
		maxRetries: maxRetries,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *WorkerPool) ProcessURLs(ctx context.Context, urls []string) chan types.WorkerResult {
	jobs := make(chan string, len(urls))
	results := make(chan types.WorkerResult, len(urls))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range jobs {
				status, err := FetchAndParseStatus(url)
				select {
				case results <- types.WorkerResult{
					URL:    url,
					Status: status,
					Error:  err,
				}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, url := range urls {
			select {
			case jobs <- url:
			case <-ctx.Done():
				return
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func (p *WorkerPool) fetchStatus(ctx context.Context, url string) (*types.Status, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/xml")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching status: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return ParseStatusXML(resp.Body)
}
