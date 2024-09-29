package pbatch_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/saenai255/pbatch"
)

func TestRun_Success(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	batchSize := 2

	process := func(n int) (string, error) {
		// Convert each number to a string
		return fmt.Sprintf("Number: %d", n), nil
	}

	results, err := pbatch.Run(items, batchSize, pbatch.STOP_ON_ERROR, process)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"Number: 1", "Number: 2", "Number: 3", "Number: 4", "Number: 5"}
	for i := range results {
		if results[i] != expected[i] {
			t.Errorf("expected result %v at index %d, got %#v", expected[i], i, results[i])
		}
	}
}

func TestRun_Error(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	batchSize := 3

	process := func(n int) (string, error) {
		// Simulate an error when processing number 3
		if n == 3 {
			return "", errors.New("error on item 3")
		}
		return fmt.Sprintf("Number: %d", n), nil
	}

	results, err := pbatch.Run(items, batchSize, pbatch.STOP_ON_ERROR, process)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "error on item 3" {
		t.Errorf("expected error 'error on item 3', got %v", err)
	}

	// Results should be nil due to the early return on error
	if results != nil {
		t.Errorf("expected results to be nil, got %v", results)
	}
}

func TestRun_EmptySlice(t *testing.T) {
	items := []int{}
	batchSize := 3

	process := func(n int) (string, error) {
		return fmt.Sprintf("Number: %d", n), nil
	}

	results, err := pbatch.Run(items, batchSize, pbatch.STOP_ON_ERROR, process)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected empty results, got %v", results)
	}
}

func TestRun_BatchSizeGreaterThanItems(t *testing.T) {
	items := []int{1, 2}
	batchSize := 10

	process := func(n int) (string, error) {
		return fmt.Sprintf("Number: %d", n), nil
	}

	results, err := pbatch.Run(items, batchSize, pbatch.STOP_ON_ERROR, process)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"Number: 1", "Number: 2"}
	for i := range results {
		if results[i] != expected[i] {
			t.Errorf("expected result %v at index %d, got %v", expected[i], i, results[i])
		}
	}
}

func TestRun_WithDelay(t *testing.T) {
	items := []int{1, 2, 3}
	batchSize := 2

	process := func(n int) (string, error) {
		time.Sleep(100 * time.Millisecond)
		return fmt.Sprintf("Number: %d", n), nil
	}

	start := time.Now()
	results, err := pbatch.Run(items, batchSize, pbatch.STOP_ON_ERROR, process)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"Number: 1", "Number: 2", "Number: 3"}
	for i := range results {
		if results[i] != expected[i] {
			t.Errorf("expected result %v at index %d, got %v", expected[i], i, results[i])
		}
	}

	if elapsed < 100*time.Millisecond || elapsed > 300*time.Millisecond {
		t.Errorf("expected processing to take around 100-300ms, but took %v", elapsed)
	}
}

func TestRun_WithDelay_Visual(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	batchSize := 2

	process := func(n int) (string, error) {
		time.Sleep(1000 * time.Millisecond)
		t.Logf("Processing item %d\n", n)
		return fmt.Sprintf("Number: %d", n), nil
	}

	start := time.Now()
	results, err := pbatch.Run(items, batchSize, pbatch.STOP_ON_ERROR, process)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"Number: 1", "Number: 2", "Number: 3", "Number: 4", "Number: 5"}
	for i := range results {
		if results[i] != expected[i] {
			t.Errorf("expected result %v at index %d, got %v", expected[i], i, results[i])
		}
	}

	if elapsed < 2500*time.Millisecond || elapsed > 3500*time.Millisecond {
		t.Errorf("expected processing to take around 2500-3500ms, but took %v", elapsed)
	}
}

func TestRun_ContinueOnErrors_Aggregate(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	batchSize := 2

	process := func(n int) (string, error) {
		// Simulate an error for odd numbers
		if n%2 != 0 {
			return "", fmt.Errorf("error processing item %d", n)
		}
		return fmt.Sprintf("Processed: %d", n), nil
	}

	results, err := pbatch.Run(items, batchSize, pbatch.CONTINUE_ON_ERROR, process)

	// Check if all items were processed
	expectedResults := []string{"", "Processed: 2", "", "Processed: 4", ""}
	for i, res := range results {
		if res != expectedResults[i] {
			t.Errorf("expected result %v at index %d, got %v", expectedResults[i], i, res)
		}
	}

	// Check if the error contains all the expected error messages
	if err == nil {
		t.Fatal("expected aggregated error, got nil")
	}

	expectedErrors := []string{
		"error processing item 1",
		"error processing item 3",
		"error processing item 5",
	}

	for _, expected := range expectedErrors {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("expected error message to contain %v, but got %v", expected, err.Error())
		}
	}
}

func TestRun_ContinueOnErrors_AllFail(t *testing.T) {
	items := []int{1, 3, 5}
	batchSize := 2

	process := func(n int) (string, error) {
		// All items will result in an error
		return "", fmt.Errorf("error processing item %d", n)
	}

	results, err := pbatch.Run(items, batchSize, pbatch.CONTINUE_ON_ERROR, process)

	// Results should all be empty strings as all items fail processing
	expectedResults := []string{"", "", ""}
	for i, res := range results {
		if res != expectedResults[i] {
			t.Errorf("expected result %v at index %d, got %v", expectedResults[i], i, res)
		}
	}

	// Check if the error contains all the expected error messages
	if err == nil {
		t.Fatal("expected aggregated error, got nil")
	}

	expectedErrors := []string{
		"error processing item 1",
		"error processing item 3",
		"error processing item 5",
	}

	for _, expected := range expectedErrors {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("expected error message to contain %v, but got %v", expected, err.Error())
		}
	}
}

func TestRun_ContinueOnErrors_SomeFail(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	batchSize := 3

	process := func(n int) (string, error) {
		// Return error for odd numbers, but process even numbers successfully
		if n%2 != 0 {
			return "", errors.New("failed on odd number")
		}
		return fmt.Sprintf("Success: %d", n), nil
	}

	results, err := pbatch.Run(items, batchSize, pbatch.CONTINUE_ON_ERROR, process)

	// Results should contain the processed results for even numbers, empty for odd numbers
	expectedResults := []string{"", "Success: 2", "", "Success: 4", ""}
	for i, res := range results {
		if res != expectedResults[i] {
			t.Errorf("expected result %v at index %d, got %v", expectedResults[i], i, res)
		}
	}

	// Check that the error contains all the errors for odd numbers
	if err == nil {
		t.Fatal("expected aggregated error, got nil")
	}

	expectedErrorMessages := []string{
		"failed on odd number",
	}

	for _, expected := range expectedErrorMessages {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("expected error message to contain %v, but got %v", expected, err.Error())
		}
	}
}

func TestRun_ContinueOnErrors_AllSuccess(t *testing.T) {
	items := []int{2, 4, 6, 8}
	batchSize := 3

	process := func(n int) (string, error) {
		return fmt.Sprintf("Success: %d", n), nil
	}

	results, err := pbatch.Run(items, batchSize, pbatch.CONTINUE_ON_ERROR, process)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedResults := []string{"Success: 2", "Success: 4", "Success: 6", "Success: 8"}
	for i, res := range results {
		if res != expectedResults[i] {
			t.Errorf("expected result %v at index %d, got %v", expectedResults[i], i, res)
		}
	}
}

func TestBatchErrorAggregation(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	batchSize := 2

	// Define a process function that will fail for odd numbers
	process := func(n int) (string, error) {
		if n%2 != 0 {
			return "", fmt.Errorf("error processing item %d", n)
		}
		return fmt.Sprintf("Processed: %d", n), nil
	}

	// Use CONTINUE_ON_ERROR to aggregate all errors
	_, err := pbatch.Run(items, batchSize, pbatch.CONTINUE_ON_ERROR, process)

	// Ensure that an error was returned
	if err == nil {
		t.Fatal("expected a BatchError, got nil")
	}

	// Check if the error is a BatchError
	if !pbatch.IsBatchError(err) {
		t.Fatalf("expected error to be a BatchError, got %T", err)
	}

	// Unwrap the BatchError to access individual errors
	errors := pbatch.UnwrapBatchError(err)

	// Expected errors for all odd items
	expectedErrors := []string{
		"error processing item 1",
		"error processing item 3",
		"error processing item 5",
	}

	// Ensure that the length of unwrapped errors matches expected errors
	if len(errors) != len(expectedErrors) {
		t.Fatalf("expected %d errors, got %d", len(expectedErrors), len(errors))
	}

	// Validate each individual error. Order is not guaranteed, so we check each error against all expected errors
	for i, individualErr := range errors {
		found := false
		for _, expected := range expectedErrors {
			if individualErr.Error() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unexpected error %q at index %d", individualErr.Error(), i)
		}
	}
}
