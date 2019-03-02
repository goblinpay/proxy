package fetcher

import (
    "net/http"
    "io/ioutil"
)

const ChunkSize = 3*1000 // read ~5kB at a time

func InitRequest(url string) (buffer []byte, err error) {
  resp, err := http.Get(url)
  if err != nil {
    return
  }
  buffer, err = ioutil.ReadAll(resp.Body)
  return
}

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