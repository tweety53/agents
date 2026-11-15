package store

import "testing"

// TestPoolConfigHonoursPoolMaxConns pins decision test-pool-size
// (KAN-1124): a DSN's pool_max_conns sizes the pool, in either DSN form;
// absent, the pool keeps _maxConns; MinConns never exceeds MaxConns; and
// jit stays off whatever the DSN says about the pool.
func TestPoolConfigHonoursPoolMaxConns(t *testing.T) {
	t.Parallel()
	const url = "postgres://u:p@localhost:5433/x?sslmode=disable"
	for _, c := range []struct {
		dsn      string
		max, min int32
	}{
		{url, 10, 2},
		{url + "&pool_max_conns=2", 2, 2},
		{url + "&pool_max_conns=1", 1, 1},
		{"host=localhost port=5433 dbname=x pool_max_conns=3", 3, 2},
	} {
		cfg, err := poolConfig(c.dsn)
		if err != nil {
			t.Fatalf("poolConfig(%q): %v", c.dsn, err)
		}
		if cfg.MaxConns != c.max || cfg.MinConns != c.min {
			t.Errorf("poolConfig(%q) = MaxConns %d MinConns %d, want %d and %d", c.dsn, cfg.MaxConns, cfg.MinConns, c.max, c.min)
		}
		if got := cfg.ConnConfig.RuntimeParams["jit"]; got != "off" {
			t.Errorf("poolConfig(%q) jit = %q, want off", c.dsn, got)
		}
	}
}
