package pbatch

import (
	"errors"
	"strings"
	"sync"
)

type errorHandler bool

const (
	STOP_ON_ERROR     errorHandler = true
	CONTINUE_ON_ERROR errorHandler = false
)

// Run runs the process function on each item in the items slice in parallel.
// Only batchSize items are processed at a time.
//
// Parameters:
//   - items: the slice of items to process
//   - batchSize: the number of items to process at a time
//   - handleErrorStrategy: whether to stop processing on the first error or continue processing. Use STOP_ON_ERROR or CONTINUE_ON_ERROR
//   - process: the function to run on each item
//
// Returns:
//   - a slice of results from the process function
//   - an error if any and handleErrorStrategy is STOP_ON_ERROR, or all errors if handleErrorStrategy is CONTINUE_ON_ERROR
func Run[T any, R any](items []T, batchSize int, handleErrorStrategy errorHandler, process func(T) (R, error)) ([]R, error) {
	// Create a channel to limit the number of concurrent goroutines
	semaphore := make(chan struct{}, batchSize)
	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	// Create a slice to store results
	results := make([]R, len(items))
	// Create an error channel to capture errors
	errChan := make(chan error, len(items))
	// Mutex to safely write results and errors
	var mu sync.Mutex
	var allErrors []error

	// Iterate over all items
	for i, item := range items {
		// Acquire a semaphore slot
		semaphore <- struct{}{}
		wg.Add(1)

		// Start a goroutine for processing the item
		go func(i int, item T) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore slot when done

			// Process the item
			result, err := process(item)
			if err != nil {
				// If handleErrorStrategy is STOP_ON_ERROR, send the first error and return early
				if handleErrorStrategy == STOP_ON_ERROR {
					select {
					case errChan <- err:
					default:
					}
					return
				}

				// Collect all errors if handleErrorStrategy is CONTINUE_ON_ERROR
				mu.Lock()
				allErrors = append(allErrors, err)
				mu.Unlock()
				return
			}

			// Store result safely
			mu.Lock()
			results[i] = result
			mu.Unlock()
		}(i, item)

		// If handleErrorStrategy is STOP_ON_ERROR, check if there's an error before continuing
		if handleErrorStrategy == STOP_ON_ERROR {
			select {
			case err := <-errChan:
				// If an error occurs, wait for all running goroutines and return early
				wg.Wait()
				return nil, err
			default:
			}
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// If stopOnError is true, check for any errors that may have occurred during processing
	if handleErrorStrategy == STOP_ON_ERROR {
		select {
		case err := <-errChan:
			return nil, err
		default:
		}
	}

	// If stopOnError is false and there are aggregated errors, return them
	if handleErrorStrategy == CONTINUE_ON_ERROR && len(allErrors) > 0 {
		return results, aggregateErrors(allErrors)
	}

	return results, nil
}

// Process runs the process function on each item in the items slice in parallel.
// It is a wrapper around Run that discards the results.
// It should be used when you only care about processing the items and not the results.
// Only batchSize items are processed at a time.
//
// Parameters:
//   - items: the slice of items to process
//   - batchSize: the number of items to process at a time
//   - handleErrorStrategy: whether to stop processing on the first error or continue processing. Use STOP_ON_ERROR or CONTINUE_ON_ERROR
//   - process: the function to run on each item
//
// Returns:
//   - an error if any and handleErrorStrategy is STOP_ON_ERROR, or all errors if handleErrorStrategy is CONTINUE_ON_ERROR
func Process[T any](items []T, batchSize int, process func(T) error) error {
	_, err := Run(items, batchSize, STOP_ON_ERROR, func(item T) (struct{}, error) {
		err := process(item)
		return struct{}{}, err
	})
	return err
}

// aggregateErrors combines multiple errors into a single error
func aggregateErrors(errs []error) error {
	var errStrings []string
	for _, err := range errs {
		errStrings = append(errStrings, err.Error())
	}
	return errors.New("multiple errors: " + strings.Join(errStrings, "; "))
}
