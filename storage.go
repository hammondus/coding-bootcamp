package main

import (
	"log"
	"os"
	"sync"
)

// saveMu serializes all on-disk saves. Several save calls are launched as
// goroutines (e.g. `go saveLessonCache()`), so without this two writers could
// interleave their output — and, more subtly, persist their snapshots out of
// order (an older snapshot marshalled first but written last would win).
// Running marshal *inside* the lock is what guarantees the last write to a
// file is also the newest state.
var saveMu sync.Mutex

// writeFileAtomic persists one shared structure to path: under saveMu it calls
// marshal to snapshot the data, writes a temp file, and renames it into place.
// The rename is atomic on the same filesystem, so a reader never observes a
// half-written file. marshal must take the structure's own lock — see
// saveUsers for the pattern.
func writeFileAtomic(path string, perm os.FileMode, marshal func() ([]byte, error)) {
	saveMu.Lock()
	defer saveMu.Unlock()

	data, err := marshal()
	if err != nil {
		log.Printf("marshal %s: %v", path, err)
		return
	}
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
