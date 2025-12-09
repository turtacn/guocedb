package diagnostic

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"
)

// Handler returns HTTP handler for diagnostic endpoints
func (d *Diagnostics) Handler() http.Handler {
	mux := http.NewServeMux()

	// Diagnostic endpoints
	mux.HandleFunc("/debug/diagnostic", d.diagnosticHandler)
	mux.HandleFunc("/debug/queries", d.activeQueriesHandler)
	mux.HandleFunc("/debug/connections", d.connectionsHandler)
	mux.HandleFunc("/debug/memory", d.memoryHandler)
	mux.HandleFunc("/debug/gc", d.gcHandler)

	// pprof endpoints
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))

	return mux
}

func (d *Diagnostics) diagnosticHandler(w http.ResponseWriter, r *http.Request) {
	info := d.Collect()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(info)
}

func (d *Diagnostics) activeQueriesHandler(w http.ResponseWriter, r *http.Request) {
	queries := d.collectQueries()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(queries)
}

func (d *Diagnostics) connectionsHandler(w http.ResponseWriter, r *http.Request) {
	conns := d.collectConnections()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conns)
}

func (d *Diagnostics) memoryHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := map[string]interface{}{
		"alloc_bytes":       m.Alloc,
		"total_alloc_bytes": m.TotalAlloc,
		"sys_bytes":         m.Sys,
		"heap_alloc_bytes":  m.HeapAlloc,
		"heap_sys_bytes":    m.HeapSys,
		"heap_idle_bytes":   m.HeapIdle,
		"heap_inuse_bytes":  m.HeapInuse,
		"heap_objects":      m.HeapObjects,
		"stack_inuse_bytes": m.StackInuse,
		"stack_sys_bytes":   m.StackSys,
		"num_gc":            m.NumGC,
		"gc_cpu_fraction":   m.GCCPUFraction,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (d *Diagnostics) gcHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Trigger GC
		runtime.GC()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"gc triggered"}`))
		return
	}

	// Return GC stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := map[string]interface{}{
		"num_gc":          m.NumGC,
		"last_gc_time":    time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
		"gc_cpu_fraction": m.GCCPUFraction,
		"pause_total_ns":  m.PauseTotalNs,
		"next_gc_bytes":   m.NextGC,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
