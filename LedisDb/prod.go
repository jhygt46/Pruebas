package main

import (
	"fmt"
	"runtime"
	"encoding/binary"
	lediscfg "github.com/ledisdb/ledisdb/config"
    "github.com/ledisdb/ledisdb/ledis"
	"github.com/valyala/fasthttp"
)

type MyHandler struct {
	Db *ledis.DB `json:"Db"`
}

func main() {

	cfg := lediscfg.NewConfigDefault()
	l, _ := ledis.Open(cfg)
	db, _ := l.Select(0)

	/*
	value := make([]byte, 5000)
	for i := uint32(0); i < 1800; i++ {
		for j := uint32(0); j < 500; j++ {
			key := append(Int32tobytes(i), Int32tobytes(j)...)
			db.Set(key, value)
		}
	}
	fmt.Println("SAVE DB")
	*/

	var port string
	if runtime.GOOS == "windows" {
		port = ":81"
	}else{
		port = ":80"
	}

	pass := &MyHandler{ Db: db }

	fasthttp.ListenAndServe(port, pass.HandleFastHTTP)

}
func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/":
			
			//cat := ParamBytes(ctx.QueryArgs().Peek("cat"))
			//cuad := ParamBytes(ctx.QueryArgs().Peek("cuad"))

			TestCat := []uint32{3, 9, 132, 189, 376, 487, 563, 612, 682, 732, 745, 795, 823, 882, 923, 945, 1012, 1087, 1123, 1129, 1221, 1276, 1387, 1412, 1445, 1498, 1512, 1572, 1676, 1704, 1734, 1765, 1787}
			TestCuad := []uint32{5, 10, 23, 45, 98, 102, 132, 143, 189, 203, 245, 289, 304, 325, 367, 398, 402, 412, 434, 456, 469, 476, 488, 493}
			
			now := time.Now()
			for i:=uint32(0); i<len(TestCat); i++ {
				for j:=uint32(0); j<len(TestCuad); j++ {

					key := append(Int32tobytes(i), Int32tobytes(j)...)
					val, _ := h.Db.Get(key)
					Silence(val)

				}
			}
			fmt.Println("time elapse:", time.Since(now))
			fmt.Fprintf(ctx, "HOLA")

		case "/insert":
			
			fmt.Fprintf(ctx, "HOLA0")
			
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}

}
func Silence(b []byte){

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
func Reverse(numbers []uint8) []uint8 {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}
func Int32tobytes(i uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return Reverse(b)
}
func check(err error) {
	if err != nil {
		fmt.Println(err)
	}
}