package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"encoding/binary"
	jsoniter "github.com/json-iterator/go"

	"github.com/dgraph-io/badger/v3"
	"github.com/valyala/fasthttp"
)

type MyHandler struct {
	Conf    Config             `json:"Conf"`
	Db      *badger.DB         `json:"Db"`
}
type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {

	db := GetDb()
	pass := &MyHandler{
		Conf:    Config{},
		Db:      db,
	}
	
	wb := db.NewWriteBatch()
	defer wb.Cancel()

	Bytes := make([]byte, 10000)
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			key := append([]byte{uint8(i)}, []byte{uint8(j)}...)
			wb.Set(key, Bytes)
		}
	}
	wb.Flush()
	fmt.Println("SAVE DB")
	

	con := context.Background()
	con, cancel := context.WithCancel(con)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		for {
			select {
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP:
					pass.Conf.init()
				case os.Interrupt:
					cancel()
					os.Exit(1)
				}
			case <-con.Done():
				log.Printf("Done.")
				os.Exit(1)
			}
		}
	}()
	go func() {
		fasthttp.ListenAndServe(":81", pass.HandleFastHTTP)
	}()
	if err := run(con, pass, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}
func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/":
			
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}

}

func GetDb() *badger.DB {
	var opts badger.Options
	if runtime.GOOS == "windows" {
		opts = badger.DefaultOptions("C:/Go/badgerDB")
	} else {
		opts = badger.DefaultOptions("/var/db")
	}
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	//defer db.Close()
	return db
}
func (h *MyHandler) SaveDb(key []uint8, value []uint8) {
	txn := h.Db.NewTransaction(true)
	err := txn.SetEntry(badger.NewEntry(key, value))
	if err != nil {
		panic(err)
	}
	err = txn.Commit()
	if err != nil {
		panic(err)
	}
}

// DAEMON //
func (h *MyHandler) StartDaemon() {
	h.Conf.Tiempo = 100 * time.Second
	fmt.Println("DAEMON")
}
func (c *Config) init() {
	var tick = flag.Duration("tick", 1*time.Second, "Ticking interval")
	c.Tiempo = *tick
}
func run(con context.Context, c *MyHandler, stdout io.Writer) error {
	c.Conf.init()
	log.SetOutput(os.Stdout)
	for {
		select {
		case <-con.Done():
			return nil
		case <-time.Tick(c.Conf.Tiempo):
			c.StartDaemon()
		}
	}
}

func Int32tobytes(i uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return Reverse(b)
}
func Reverse(numbers []uint8) []uint8 {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}
func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}