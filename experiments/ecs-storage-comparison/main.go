package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"ecs-storage-comparison/archetype"
	"ecs-storage-comparison/storage"
	"ecs-storage-comparison/workload"
)

func main() {
	var (
		backend  = flag.String("backend", "archetype", "storage backend")
		workload = flag.String("workload", "iteration", "workload name")
		scale    = flag.Int("scale", 1000, "entity count")
		frames   = flag.Int("frames", 1000, "frame count")
		out      = flag.String("out", "results", "output directory")
	)
	flag.Parse()

	s := newBackend(*backend)
	setup, tick := selectWorkload(*workload)
	setup(s, *scale)

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	times := make([]int64, *frames)
	for i := range *frames {
		start := time.Now()
		tick(s)
		times[i] = time.Since(start).Nanoseconds()
	}

	runtime.ReadMemStats(&after)

	if err := writeCSV(*out, *backend, *workload, *scale, times, before, after); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newBackend(name string) storage.Storage {
	switch name {
	case "archetype":
		return archetype.New()
	default:
		fmt.Fprintf(os.Stderr, "unknown backend: %s\n", name)
		os.Exit(1)
		return nil
	}
}

func selectWorkload(name string) (func(storage.Storage, int), func(storage.Storage)) {
	switch name {
	case "iteration":
		return workload.IterationSetup, workload.IterationTick
	default:
		fmt.Fprintf(os.Stderr, "unknown workload: %s\n", name)
		os.Exit(1)
		return nil, nil
	}
}

func writeCSV(dir, backend, wlName string, scale int, times []int64, before, after runtime.MemStats) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s_%s_%d.csv", dir, backend, wlName, scale)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"frame", "time_ns"}); err != nil {
		return err
	}
	for i, t := range times {
		if err := w.Write([]string{strconv.Itoa(i), strconv.FormatInt(t, 10)}); err != nil {
			return err
		}
	}

	frames := len(times)
	allocsPerFrame := float64(after.Mallocs-before.Mallocs) / float64(frames)
	bytesPerFrame := float64(after.TotalAlloc-before.TotalAlloc) / float64(frames)
	fmt.Printf("wrote %s\n", path)
	fmt.Printf(
		"  frames=%d allocs/frame=%.2f bytes/frame=%.2f peak_heap=%d\n",
		frames, allocsPerFrame, bytesPerFrame, after.HeapInuse,
	)

	return nil
}
