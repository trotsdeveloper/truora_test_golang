package main

import (
  "database/sql"
  "fmt"
  "log"
  _ "github.com/lib/pq"
)

func main() {
  // Connect to the "servers" database.
  db, err := sql.Open("postgres",
      "postgresql://manuelams@localhost:26257/bank?ssl=true&sslmode=require&sslrootcert=../certs/ca.crt&sslkey=../certs/client.maxroach.key&sslcert=../certs/client.maxroach.crt")
  if err != nil {
      log.Fatal("error connecting to the database: ", err)
  }
  fmt.Println("Hello world")
  defer db.Close()

}
