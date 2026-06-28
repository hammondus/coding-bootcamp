package main

import (
	"log"
	"os"
	"sync"
)

// saveMu serializes all on-disk writes. Several save calls are launched as
// goroutines (e.g. `go saveLessonCache()`), so without this two writers could
// hit the same file concurrently and interleave their output.
var saveMu sync.Mutex

// writeFileAtomic writes data to path by writing a temp file and renaming it
// into place. The rename is atomic on the same filesystem, so a reader never
// observes a half-written file, and saveMu guarantees writers don't overlap.
func writeFileAtomic(path string, data []byte, perm os.FileMode) {
	saveMu.Lock()
	defer saveMu.Unlock()

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		log.Printf("write %s: %v", tmp, err)
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		log.Printf("rename %s -> %s: %v", tmp, path, err)
		os.Remove(tmp)
	}
}
