package main

import (
	"fmt"
	"iter"
)

type StreamX[T any] iter.Seq[T]
type StreamXMapper[Input any, Output any] func(input StreamX[Input]) StreamX[Output]

// SliceToStream converts a slice into a StreamX iterator.
func SliceToStream[T any](inputSlice []T) StreamX[T] {
	return func(yield func(T) bool) {
		for _, item := range inputSlice {
			if !yield(item) {
				return // Stop iteration if yield returns false
			}
		}
	}
}

func StreamToSlice[Type any](stream StreamX[Type]) []Type {
	result := make([]Type, 0)
	for item := range stream {
		result = append(result, item)
	}
	return result
}

func Map[Input, Output any](mapper func(Input) Output) StreamXMapper[Input, Output] {
	return func(inputStream StreamX[Input]) StreamX[Output] {
		// returns the output stream
		return func(yield func(Output) bool) {
			// Consume the input stream, apply the transformation, and yield the result
			inputStream(func(val Input) bool {
				return yield(mapper(val))
			})
		}
	}
}

func Tap[Input, Output any](tapper func(Input) Output) StreamXMapper[Input, Input] {
	return func(inputStream StreamX[Input]) StreamX[Input] {
		// returns the output stream
		return func(yield func(Input) bool) {
			// Consume the input stream, apply the transformation, and yield the result
			inputStream(func(val Input) bool {
				tapper(val)
				return yield(val)
			})
		}
	}
}

// Filter filters the stream based on a condition.
func Filter[Input any](condition func(Input) bool) StreamXMapper[Input, Input] {
	return func(inputStream StreamX[Input]) StreamX[Input] {
		return func(yield func(Input) bool) {
			// Consume the input stream and apply the filter condition
			inputStream(func(val Input) bool {
				// Only yield values that satisfy the condition
				if condition(val) {
					return yield(val)
				}
				return true // Continue iterating
			})
		}
	}
}

func Batch[Input any](size int) StreamXMapper[Input, []Input] {
	return func(inputStream StreamX[Input]) StreamX[[]Input] {
		return func(yield func([]Input) bool) {
			var batched []Input
			// Consume the input stream and accumulate items into batches
			inputStream(func(val Input) bool {
				batched = append(batched, val)

				// If the batch reaches the specified size, yield it
				if len(batched) >= size {
					toEmit := batched
					batched = nil // Reset the batch
					return yield(toEmit)
				}
				return true // Continue iterating
			})

			// If there are remaining items in the batch, yield them
			if len(batched) > 0 {
				toEmit := batched
				batched = nil // Reset the batch
				yield(toEmit)
			}
		}
	}
}

func Flat[Output any]() StreamXMapper[[]Output, Output] {
	return func(inputStream StreamX[[]Output]) StreamX[Output] {
		return func(yield func(Output) bool) {
			// Consume the input stream and accumulate items into batches
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

func Run[Type any](stream StreamX[Type]) bool {
	for _ = range stream {
		// do nothing
		continue
	}

	return true
}

func Pipeline[Input any](mappers ...StreamXMapper[Input, Input]) StreamXMapper[Input, Input] {
	return func(inputStream StreamX[Input]) StreamX[Input] {
		for _, mapper := range mappers {
			inputStream = mapper(inputStream)
		}

		return inputStream
	}
}

func Pipe[Input any, Output any](input StreamX[Input], mapper StreamXMapper[Input, Output], mappers ...StreamXMapper[Output, Output]) StreamX[Output] {
	output := mapper(input)
	for _, mapper := range mappers {
		output = mapper(output)
	}
	return output
}

func main() {
	// Create a StreamX From the input slice
	stream0 := SliceToStream([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})

	myIntPipeline := Pipeline[int](
		Tap(func(input int) any {
			fmt.Println(input)
			return nil
		}),
		Filter(func(input int) bool {
			return input > 3
		}),
		Map(func(input int) int {
			return input + 100
		}),
		Tap(func(input int) any {
			fmt.Println(input)
			return nil
		}),
	)

	processedStream := Pipe(stream0, myIntPipeline)
	batchedResultStream := Pipe(processedStream, Batch[int](3))

	//final := logValue(flatten(logBatched(batchOutput(toValue(filter1(logInputStream(stream0)))))))

	log1 := Log[[]int]("Batched")
	batchedResultStreamWithLogs := Pipe(batchedResultStream, log1)

	log2 := Log[int]("Flatten")

	final := Pipe(batchedResultStreamWithLogs, Flat[int](), log2)

	fmt.Println(StreamToSlice(final))

}

func Log[Type any](label string) StreamXMapper[Type, Type] {
	return Tap(func(input Type) any {
		fmt.Println(label, input)
		return nil
	})
}
