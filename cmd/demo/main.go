package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Clint-Mathews/fencelock/etcdlock"
	"github.com/Clint-Mathews/fencelock/fencedstore"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	endpoint := flag.String("endpoint", "localhost:2379", "etcd endpoint")
	flag.Parse()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{*endpoint},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	locker := etcdlock.New(cli)
	resouce := fencedstore.NewMemory()
	ctx := context.Background()

	fmt.Println(" ==== Client A acquires ====")
	leaseA, err := locker.Acquire(ctx, "demo-resource", 2*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("A holds lease, token=%d\n", leaseA.Token)
	if err := resouce.Write(ctx, "demo-resource", leaseA.Token, []byte("A's data")); err != nil {
		log.Fatal(err)
	}
	fmt.Println("A wrote successfully")

	fmt.Println("\n==== Client A pauses (simulated GC stall, 3s > 2s TTL) ====")
	// Sleeping is not enough against etcd: this process's Session Keepalive
	// is still running.
	_ = cli.Close()
	time.Sleep(3 * time.Second)

	cliB, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{*endpoint},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cliB.Close()
	locker = etcdlock.New(cliB)

	fmt.Println("\n==== Client B acquires (A's lease has expired) ====")
	leaseB, err := locker.Acquire(ctx, "demo-resource", 1*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("B holds lease, token=%d (> A's token %d)\n", leaseB.Token, leaseA.Token)
	if err := resouce.Write(ctx, "demo-resource", leaseB.Token, []byte("B's data")); err != nil {
		log.Fatal(err)
	}
	fmt.Println("B wrote successfully")

	fmt.Println("\n==== Client A wakes up, tries to write its stale token ====")
	if err := resouce.Write(ctx, "demo-resource", leaseA.Token, []byte("A's stale data")); err != nil {
		fmt.Printf("REJECTED as expected: %v \n", err)
	} else {
		fmt.Println("!! BUG: Stale write was not rejected")
	}
	fmt.Printf("\nFinal resource state: %q\n", resouce.Get("demo-resource"))
}
