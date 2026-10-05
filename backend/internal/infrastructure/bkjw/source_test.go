//go:build test

package bkjw

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestCollectorProtocolAndCancellation(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node runtime unavailable")
	}
	file := filepath.Join(t.TempDir(), "runner.js")
	if err := os.WriteFile(file, []byte(`const input = require('node:readline').createInterface({input: process.stdin});
console.log(JSON.stringify({type:'ready',semesters:['2026-2027-1']}));
input.on('line', line => {const cmd = JSON.parse(line); if(cmd.type === 'sync') { console.log(JSON.stringify({type:'collecting',semester:cmd.semesters[0]})); console.log(JSON.stringify({type:'progress',semester:cmd.semesters[0],progress:{stage:'details',completed:3,total:10,failed:1,retries:2}})); } else if(cmd.type === 'cancel') input.close();});
input.on('close', () => process.exit(0));`), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn, err := (&Source{Config: Config{Enabled: true, Runner: file}}).Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	select {
	case event := <-conn.Events():
		if event.Type != "ready" {
			t.Fatal(event)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no ready event")
	}
	if err := conn.Sync([]string{"2026-2027-1"}); err != nil {
		t.Fatal(err)
	}
	select {
	case event := <-conn.Events():
		if event.Semester != "2026-2027-1" {
			t.Fatal(event)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no collecting event")
	}
	select {
	case event := <-conn.Events():
		if event.Progress == nil || event.Progress.Stage != "details" || event.Progress.Completed != 3 || event.Progress.Total == nil || *event.Progress.Total != 10 || event.Progress.Failed != 1 || event.Progress.Retries != 2 {
			t.Fatalf("lost collector progress: %+v", event)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no detailed progress event")
	}
	conn.Close()
	select {
	case <-conn.(*connection).done:
	case <-time.After(3 * time.Second):
		t.Fatal("collector did not shut down")
	}
}
