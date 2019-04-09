package fetcher

import (
    "net/http"
    "io/ioutil"
    "github.com/gregjones/httpcache" // check perf/doc, use string keys which cause GC overhead
)

const ChunkSize = 3*1000 // read ~5kB at a time

var tp = httpcache.NewMemoryCacheTransport()
var client = &http.Client{Transport: tp}

func Fetch(url string) (body []byte, err error) {
  resp, err := client.Get(url)
  if err != nil {
    // handle error
    return
  }
  defer resp.Body.Close()
  body, err = ioutil.ReadAll(resp.Body) // anything more efficient that avoids a copy if cached?
  return
}

// TODO: look into using bufio or bytes.Buffer
func ReadChunk(buffer *[]byte) string {
  b := *buffer
  s := ChunkSize
  if len(b) < s {
    s = len(b)
  }

  c := string(b[:s])
  *buffer = b[s:]
  return c
}



// func InitRequest(url string) (resp *http.Response, err error) {
//   resp, err = http.Get(url)
//   return
// }

// func ReadChunk(rc io.ReadCloser) (c string, err error) {

//   b := make([]byte, ChunkSize)
//   n, err := rc.Read(b) // how to read to string directly?

//   if n < ChunkSize || err == io.EOF {
//     err = errors.New("Done reading.")
//     //rc.Close() // really needed?
//   }

//   c = string(b[:n])

//   return
// }
