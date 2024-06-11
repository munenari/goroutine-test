package main

import (
	"runtime"
	"sync"
	"testing"

	"golang.org/x/sync/errgroup"
)

func TestJob(t *testing.T) {
	t.Run("small-and-light", func(t *testing.T) {
		x := &worker{
			pool: smallPool,
			job:  lightJob,
		}
		if err := x.Do(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("big-and-heavy", func(t *testing.T) {
		x := &worker{
			pool: bigPool,
			job:  heavyJob,
		}
		if err := x.Do(); err != nil {
			t.Fatal(err)
		}
	})
}

func benchmarkLoop(b *testing.B, x *worker) {
	b.Run("for", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := x.Do(); err != nil {
				b.Error(err)
			}
		}
	})
	b.Run("go", func(b *testing.B) {
		wg := &sync.WaitGroup{}
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := x.Do(); err != nil {
					b.Error(err)
				}
			}()
		}
		wg.Wait()
	})
	b.Run("errgroup-nolimit", func(b *testing.B) {
		eg := &errgroup.Group{}
		for i := 0; i < b.N; i++ {
			eg.Go(func() error {
				return x.Do()
			})
		}
		if err := eg.Wait(); err != nil {
			b.Error(err)
		}
	})
	b.Run("errgroup", func(b *testing.B) {
		eg := &errgroup.Group{}
		eg.SetLimit(runtime.NumCPU())
		for i := 0; i < b.N; i++ {
			eg.Go(func() error {
				return x.Do()
			})
		}
		if err := eg.Wait(); err != nil {
			b.Error(err)
		}
	})
	b.Run("threadworker", func(b *testing.B) {
		workerNum := runtime.NumCPU()
		wg := &sync.WaitGroup{}
		wg.Add(workerNum)
		q := make(chan struct{}, workerNum*100)
		for w := 0; w < workerNum; w++ {
			go func() {
				defer wg.Done()
				for range q {
					if err := x.Do(); err != nil {
						b.Error(err)
					}
				}
			}()
		}
		for i := 0; i < b.N; i++ {
			q <- struct{}{}
		}
		close(q)
		wg.Wait()
	})
}

func BenchmarkJob(b *testing.B) {
	b.Run("small-light", func(b *testing.B) {
		benchmarkLoop(b, &worker{
			pool: smallPool,
			job:  lightJob,
		})
	})
	b.Run("small-heavy", func(b *testing.B) {
		benchmarkLoop(b, &worker{
			pool: smallPool,
			job:  heavyJob,
		})
	})
	b.Run("big-light", func(b *testing.B) {
		benchmarkLoop(b, &worker{
			pool: bigPool,
			job:  lightJob,
		})
	})
	b.Run("big-heavy", func(b *testing.B) {
		benchmarkLoop(b, &worker{
			pool: bigPool,
			job:  heavyJob,
		})
	})
}
