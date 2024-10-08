# pbatch

`pbatch` is a lightweight Go package for processing a slice of items concurrently, with control over error handling. You can control how many items are processed at a time (`batchSize`) and decide whether to stop processing upon encountering an error or continue processing all items while aggregating errors.

## Features

- **Concurrent Batch Processing**: Process items concurrently with a configurable batch size.
- **Customizable Error Handling**: Option to stop on the first error (`STOP_ON_ERROR`) or continue processing all items and aggregate errors (`CONTINUE_ON_ERROR`).
- **Flexible Processing Functions**: Supports functions that return results or simply perform actions on items.
- **Batch Error Aggregation**: When using `CONTINUE_ON_ERROR`, errors are aggregated into a custom `BatchError` type.

## Installation

To install `pbatch`, use:

```bash
go get github.com/saenai255/pbatch
```

## Usage

### Importing the Package

```go
import "github.com/saenai255/pbatch"
```

### `Run` Function

The main function in the package is `Run`, which processes items concurrently with a specified batch size and error-handling strategy.

#### Signature

```go
func Run[T any, R any](items []T, batchSize int, handleErrorStrategy errorHandler, process func(T) (R, error)) ([]R, error)
```

#### Parameters

- **`items`**: A slice of items to process.
- **`batchSize`**: The maximum number of items to process concurrently.
- **`handleErrorStrategy`**: Error handling strategy — either `STOP_ON_ERROR` or `CONTINUE_ON_ERROR`.
- **`process`**: A function that takes an item and returns a result and an error.

#### Returns

- **`[]R`**: A slice of results from the `process` function.
- **`error`**: 
  - If `handleErrorStrategy` is `STOP_ON_ERROR`, an error is returned when any error occurs, stopping further processing.
  - If `handleErrorStrategy` is `CONTINUE_ON_ERROR`, a `BatchError` containing all errors is returned.

### Example: Using `Run`

```go
package main

import (
	"fmt"
	"github.com/saenai255/pbatch"
)

func main() {
	items := []int{1, 2, 3, 4, 5}
	batchSize := 2

	// Process function that returns the square of the number or an error
	processFunc := func(n int) (int, error) {
		if n == 3 {
			return 0, fmt.Errorf("error processing item %d", n)
		}
		return n * n, nil
	}

	// Run with STOP_ON_ERROR
	results, err := pbatch.Run(items, batchSize, pbatch.STOP_ON_ERROR, processFunc)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Results:", results)
	}

	// Run with CONTINUE_ON_ERROR
	results, err = pbatch.Run(items, batchSize, pbatch.CONTINUE_ON_ERROR, processFunc)
	if err != nil {
		fmt.Println("BatchError:", err)
		if pbatch.IsBatchError(err) {
			for _, individualErr := range pbatch.UnwrapBatchError(err) {
				fmt.Println(" -", individualErr)
			}
		}
	}
	fmt.Println("Results:", results)
}
```

#### Example Output

```
Error: error processing item 3
Results: []

BatchError: batch error: error processing item 3
 - error processing item 3
Results: [1 4 0 16 25]
```

### Error Handling Strategies

- **`STOP_ON_ERROR`**: Processing stops as soon as an error is encountered, and that error is returned.
- **`CONTINUE_ON_ERROR`**: Processing continues for all items, and any errors encountered are aggregated into a `BatchError`.

### `Process` Function

The `Process` function is a wrapper around `Run` for cases where you don't need the results and only care about processing items.

#### Signature

```go
func Process[T any](items []T, batchSize int, process func(T) error) error
```

#### Parameters

- **`items`**: A slice of items to process.
- **`batchSize`**: The number of items to process concurrently.
- **`process`**: A function to perform an operation on each item, returning an error if any occurs.

#### Returns

- **`error`**: 
  - An error if any occur and `handleErrorStrategy` is `STOP_ON_ERROR`.
  - A `BatchError` containing all errors if `handleErrorStrategy` is `CONTINUE_ON_ERROR`.

#### Example: Using `Process`

```go
package main

import (
	"fmt"
	"github.com/saenai255/pbatch"
)

func main() {
	items := []int{1, 2, 3, 4, 5}
	batchSize := 2

	// Process function that prints the number or returns an error
	processFunc := func(n int) error {
		if n == 3 {
			return fmt.Errorf("error processing item %d", n)
		}
		fmt.Println("Processed:", n)
		return nil
	}

	// Process items with STOP_ON_ERROR
	err := pbatch.Process(items, batchSize, processFunc)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
```

### `BatchError` Type

`pbatch` introduces a `BatchError` type to handle scenarios where multiple errors occur during processing. This allows all errors to be captured and returned as a single entity.

#### `BatchError` Functions
- **`Error() string`**: Returns a concatenated string representation of all contained errors.
- **`IsBatchError(error)`**: Checks if an error is a `BatchError`.
- **`UnwrapBatchError(error)`**: Extracts all errors from a `BatchError`. If the error is not a `BatchError`, it returns the error in a slice.

#### Example: Working with `BatchError`

```go
err := pbatch.Run(items, batchSize, pbatch.CONTINUE_ON_ERROR, processFunc)
if err != nil && pbatch.IsBatchError(err) {
    errors := pbatch.UnwrapBatchError(err)
    for _, e := range errors {
        fmt.Println("Individual error:", e)
    }
}
```

### Error Aggregation

When using `CONTINUE_ON_ERROR`, any errors encountered are combined into a `BatchError`. This aggregated error contains all individual errors as a slice, allowing for easy inspection and handling.

### License

This library is licensed under the MIT License.

---

With `pbatch`, you can easily process slices of items concurrently, with control over the batch size and error-handling strategy. Use `STOP_ON_ERROR` for immediate error stops, or `CONTINUE_ON_ERROR` to aggregate errors and continue processing seamlessly.