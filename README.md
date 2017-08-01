### RUN:

PAYLOAD='[
  {
    "method": "GET",
    "url": "http://httpbin.org/delay/1",
    "headers": {
      "User-Agent": "Golang tester 1",
      "Content-Type": "text/html"
    },
    "body": "nothing"
  },
  {
    "method": "GET",
    "url": "http://httpbin.org/delay/3",
    "headers": {
      "User-Agent": "Golang tester 3",
      "Content-Type": "text/html"
    },
    "body": "nothing"
  }
]' go run httpclient.go