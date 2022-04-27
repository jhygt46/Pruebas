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

	var port string
	if runtime.GOOS == "windows" {
		port = ":81"
	}else{
		port = ":80"
	}

	db := GetDb()
	pass := &MyHandler{
		Conf:    Config{},
		Db:      db,
	}

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
		fasthttp.ListenAndServe(port, pass.HandleFastHTTP)
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
			
			err := h.Db.View(func(txn *badger.Txn) error {

				cat := ParamBytes(ctx.QueryArgs().Peek("cat"))
				cuad := ParamBytes(ctx.QueryArgs().Peek("cuad"))

				Key := append(cat, cuad...)
				item, err := txn.Get(Key)
				if err == nil {
					val, err := item.ValueCopy(nil)
					if err == nil {
						fmt.Println(val)
					}
				}
				return nil

			})
			check(err)
			fmt.Fprintf(ctx, "HOLA0")
			
		case "/insert1":

			istart := ParamUint32(ctx.QueryArgs().Peek("istart"))
			jstart := ParamUint32(ctx.QueryArgs().Peek("jstart"))

			iend := ParamUint32(ctx.QueryArgs().Peek("iend"))
			jend := ParamUint32(ctx.QueryArgs().Peek("jend"))

			Bytes := make([]byte, 10000)

			for i := istart; i < iend; i++ {
				for j := jstart; j < jend; j++ {
					key := append(Int32tobytes(i), Int32tobytes(j)...)
					h.SaveDb(key, Bytes)
				}
			}
			fmt.Fprintf(ctx, "HOLA1")

		case "/insert2":

			istart := ParamUint32(ctx.QueryArgs().Peek("istart"))
			jstart := ParamUint32(ctx.QueryArgs().Peek("jstart"))

			iend := ParamUint32(ctx.QueryArgs().Peek("iend"))
			jend := ParamUint32(ctx.QueryArgs().Peek("jend"))

			Bytes := make([]byte, 10000)

			for i := istart; i < iend; i++ {
				for j := jstart; j < jend; j++ {
					key := append(Int32tobytes(i), Int32tobytes(j)...)
					h.SaveDb2(key, Bytes)
				}
			}
			fmt.Fprintf(ctx, "HOLA2")

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
	opts.SyncWrites = false
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	return db
}
func (h *MyHandler) SaveDb2(key []byte, value []byte){
	wb := h.Db.NewWriteBatch()
	defer wb.Cancel()
	check(wb.Set(key, value))
	check(wb.Flush())
}
func (h *MyHandler) SaveDb(key []byte, value []byte) {
	txn := h.Db.NewTransaction(true)
	check(txn.SetEntry(badger.NewEntry(key, value)))
	check(txn.Commit())
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
func ParamUint32(data []byte) uint32 {
	var x uint32
	for _, c := range data {
		x = x*10 + uint32(c-'0')
	}
	return x
}
func ParamBytes(data []byte) []byte {
	var x uint32
	for _, c := range data {
		x = x*10 + uint32(c-'0')
	}
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, x)
	return Reverse(b)
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