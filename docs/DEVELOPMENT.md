# whimer

Project structure rule

```
.
├── cmd
│   └── main.go     // main entry
├── etc
│   └── xxx.yaml    // configuration file
├── go.mod
├── go.sum
└── internal
    ├── config      // configuration definition
    ├── global      // global variables or constants
    ├── handler     // http routes and its handler
    ├── model       // model used inside project
    ├── repo        // database 
    ├── svc         // service implementation
    └── types       // model used by api (http)

```