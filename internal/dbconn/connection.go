// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbconn

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lib/pq"
)

// Allow lib/pq to send CancelRequest before forcing a stalled connection closed.
const cancellationGracePeriod = time.Second

// Open binds database connections to ctx, including connection startup and
// reads waiting for lost responses.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	connector, err := pq.NewConnector(dsn)
	if err != nil {
		return nil, err
	}
	connector.Dialer(&contextDialer{})
	return sql.OpenDB(contextConnector{Connector: connector, ctx: ctx}), nil
}

type connectionContextKey struct{}

type contextConnector struct {
	*pq.Connector
	ctx context.Context
}

func (c contextConnector) Connect(ctx context.Context) (driver.Conn, error) {
	connectCtx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(c.ctx, cancel)
	defer stop()
	defer cancel()
	return c.Connector.Connect(context.WithValue(connectCtx,
		connectionContextKey{}, c.ctx))
}

type contextDialer struct {
	net.Dialer
}

func (d *contextDialer) DialTimeout(network, address string,
	timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return d.DialContext(ctx, network, address)
}

func (d *contextDialer) DialContext(ctx context.Context,
	network, address string) (net.Conn, error) {
	conn, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	// Regular connections follow the scrape's lifetime, not the driver's
	// short-lived connect_timeout context. CancelRequest uses its own context
	// so it can still reach the server after the scrape has been canceled.
	lifetime, ok := ctx.Value(connectionContextKey{}).(context.Context)
	if !ok {
		lifetime = ctx
	}
	c := &contextConn{Conn: conn, ctx: lifetime, closed: make(chan struct{})}
	c.stop = context.AfterFunc(lifetime, func() {
		timer := time.NewTimer(cancellationGracePeriod)
		defer timer.Stop()
		select {
		case <-timer.C:
			_ = abortConnection(conn)
		case <-c.closed:
		}
	})
	return c, nil
}

type contextConn struct {
	net.Conn
	ctx    context.Context
	closed chan struct{}
	stop   func() bool
	once   sync.Once
	failed atomic.Bool
}

func (c *contextConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if err != nil {
		c.failed.Store(true)
	}
	return n, err
}

func (c *contextConn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	if err != nil {
		c.failed.Store(true)
	}
	return n, err
}

func (c *contextConn) Close() error {
	c.once.Do(func() {
		c.stop()
		close(c.closed)
	})
	if c.ctx.Err() != nil || c.failed.Load() {
		return abortConnection(c.Conn)
	}
	return c.Conn.Close()
}

func abortConnection(conn net.Conn) error {
	// Do not leave TCP waiting for acknowledgments of lost data and FIN.
	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.SetLinger(0)
	}
	return conn.Close()
}
