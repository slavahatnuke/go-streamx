# go-streamx

Motivation to create some toolkit to manage Go iterators.

It is a quite useful to use streams and iterators are streams by its nature. 
Its getting handy in case when you need to handle big volumes of data to map it, transform, validate, filter and so on.

- Go iterators https://pkg.go.dev/iter are quite performant
- It's possible to have strictly typed mappers
- Filters
- Map-Reduce for analytics
- In stream transformations
- Allows to make scalability with go routines

Conceptually looks a quite good to work with bigger data.

# Concept

The concept you've developed with `go-streamx` is an extension of the iterator pattern in Go. It provides a **functional, pipeline-based approach** to handle streams of data, enabling operations such as **mapping**, **filtering**, **batching**, and **flattening** of data in a **lazy** and **efficient** manner. This approach is well-suited for processing large datasets, where you need to perform operations like transformation or validation without loading everything into memory at once.

### Key Concepts in `go-streamx`

1. **Streams and Iterators**:
    - **Streams** in `go-streamx` represent sequences of data that are processed one element at a time.
    - You use iterators to lazily retrieve values from a sequence, ensuring that only the required data is computed and held in memory at any given time.

2. **Mappers**:
    - The **`StreamXMapper`** type represents functions that transform one stream into another. These mappers are essentially higher-order functions that allow for transformations like **map**, **filter**, **batching**, and others in a lazy, chained manner.

3. **Lazy Evaluation**:
    - Since the data is processed one item at a time, the code avoids memory overhead and allows for **on-demand processing**. This is particularly beneficial when handling large datasets where you don't want to process everything at once.

4. **Stream Operations**:
    - The library defines several operations like **Map**, **Filter**, **Batch**, and **Flat** that operate on streams to modify or retrieve data.
    - **Map** transforms elements, **Filter** allows for conditional data extraction, **Batch** groups elements into batches, and **Flat** flattens nested structures.

### Example Breakdown

Here is a breakdown of some core operations and their usage:

#### 1. **Slice to Stream**
This converts a slice into a stream, allowing for lazy iteration over its elements.

```go
func SliceToStream[T any](inputSlice []T) StreamX[T] {
	return func(yield func(T) bool) {
		for _, item := range inputSlice {
			if !yield(item) {
				return
			}
		}
	}
}
```

#### 2. **Map Transformation**
The **`Map`** function applies a transformation to each item in the stream.

```go
func Map[Input, Output any](mapper func(Input) Output) StreamXMapper[Input, Output] {
	return func(inputStream StreamX[Input]) StreamX[Output] {
		return func(yield func(Output) bool) {
			inputStream(func(val Input) bool {
				return yield(mapper(val))
			})
		}
	}
}
```
- This allows you to map over the stream and apply transformations.

#### 3. **Filter**
The **`Filter`** function allows you to selectively pass only elements that satisfy a given condition.

```go
func Filter[Input any](condition func(Input) bool) StreamXMapper[Input, Input] {
	return func(inputStream StreamX[Input]) StreamX[Input] {
		return func(yield func(Input) bool) {
			inputStream(func(val Input) bool {
				if condition(val) {
					return yield(val)
				}
				return true // Continue iterating
			})
		}
	}
}
```
- It ensures that only elements meeting the condition are passed down the pipeline.

#### 4. **Batching**
The **`Batch`** operation groups elements into batches of a specified size.

```go
func Batch[Input any](size int) StreamXMapper[Input, []Input] {
	return func(inputStream StreamX[Input]) StreamX[[]Input] {
		return func(yield func([]Input) bool) {
			var batched []Input
			inputStream(func(val Input) bool {
				batched = append(batched, val)

				if len(batched) >= size {
					toEmit := batched
					batched = nil // Reset the batch
					return yield(toEmit)
				}
				return true
			})

			if len(batched) > 0 {
				toEmit := batched
				batched = nil
				yield(toEmit)
			}
		}
	}
}
```
- This is useful when you want to process data in chunks, which is particularly important for large datasets or API limits.

#### 5. **Flat**
The **`Flat`** function flattens nested structures (arrays of arrays).

```go
func Flat[Output any]() StreamXMapper[[]Output, Output] {
	return func(inputStream StreamX[[]Output]) StreamX[Output] {
		return func(yield func(Output) bool) {
			inputStream(func(val []Output) bool {
				for _, item := range val {
					if !yield(item) {
						return false
					}
				}
				return true
			})
		}
	}
}
```
- This is handy when you're working with collections that contain other collections, and you want to simplify the structure for further processing.

#### 6. **Pipelines**
The power of `go-streamx` comes from combining these operations into **pipelines**, where each stage of the pipeline transforms or filters the data.

```go
func Pipeline[Input any](mappers ...StreamXMapper[Input, Input]) StreamXMapper[Input, Input] {
	return func(inputStream StreamX[Input]) StreamX[Input] {
		for _, mapper := range mappers {
			inputStream = mapper(inputStream)
		}
		return inputStream
	}
}
```
- You can chain together multiple mappers like **Filter**, **Map**, **Batch**, etc., and the resulting stream will process data through all the stages.

### Example of Usage

```go
func main() {
	stream0 := SliceToStream([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})

	myIntPipeline := Pipeline[int](
		Filter(func(input int) bool { return input > 3 }),
		Map(func(input int) int { return input + 100 }),
	)

	processedStream := Pipe(stream0, myIntPipeline)
	batchedResultStream := Pipe(processedStream, Batch )

	// Log outputs
	log1 := Log[[]int]("Batched")
	batchedResultStreamWithLogs := Pipe(batchedResultStream, log1)

	log2 := Log[int]("Flatten")
	final := Pipe(batchedResultStreamWithLogs, Flat[int](), log2)

	fmt.Println(StreamToSlice(final))
}
```
- In this example, we process a stream of integers: we first **filter** values greater than 3, then **map** each value by adding 100, **batch** them into groups of 3, and finally **flatten** the batches before printing.

### Motivation and Advantages

- **Memory Efficiency**: The iterator pattern ensures that data is not loaded into memory all at once. Data is processed lazily, improving memory usage for large datasets.
- **Composable Operations**: You can compose complex operations (filter, map, batch, etc.) in a functional pipeline style, which simplifies your code and makes it more readable.
- **Error Handling**: By yielding both data and errors in an iterator, you can gracefully handle failures and stop processing as needed.
- **Concurrency**: Go's concurrency model can be used to scale stream processing, allowing parallel processing of different stream operations in separate goroutines.

### Conclusion

The `go-streamx` library enhances Go's iterator pattern by providing a functional and composable toolkit for stream processing. It’s an ideal choice when working with large datasets, offering both **memory efficiency** and **transformational flexibility**. This functional approach allows you to easily create **pipelines** of operations, process data lazily, and scale with concurrency when necessary.
