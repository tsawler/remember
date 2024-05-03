<a href="https://golang.org"><img src="https://img.shields.io/badge/powered_by-Go-3362c2.svg?style=flat-square" alt="Built with GoLang"></a>
[![Version](https://img.shields.io/badge/goversion-1.20.x-blue.svg)](https://golang.org)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/tsawler/remember/master/LICENSE.md)
<a href="https://pkg.go.dev/github.com/tsawler/remember"><img src="https://img.shields.io/badge/godoc-reference-%23007d9c.svg"></a>
[![Go Report Card](https://goreportcard.com/badge/github.com/tsawler/remember)](https://goreportcard.com/report/github.com/tsawler/remember)
![Tests](https://github.com/tsawler/remember/actions/workflows/tests.yml/badge.svg)

# Remember

Package remember provides an easy way to implement a Redis cache in your Go application. 

## Installation
Install it in the usual way:

`go get -u github.com/tsawler/remember`

# Usage
Create an instance of the `remember.Cache` type by using the `remember.New()` function, and optionally
passing it a `remember.Options` variable:

~~~go
cache := remember.New() // Will use default options, suitable for development.
~~~

Or, specifying options:
~~~go
ops := remember.Options{
    Server:   "localhost"      // The server where Redis exists.
    Port:     "6379"           // The port Redis is listening on.
    Password: "some_password"  // The password for Redis.
    Prefix:   "myapp"          // A prefix to use for all keys for this client. Useful when multiple clients use the same database.
    DB:       1                // Database. Specifying 0 (the default) means use the default database.
}

cache := remember.New(ops)
~~~

## Example Program

~~~go
package main

import (
	"encoding/gob"
	"fmt"
	"github.com/tsawler/remember"
	"log"
	"os"
	"time"
)

type Student struct {
	Name string
	Age  int
}

func main() {
	// For non-scalar types, you must register the type using gob.Register.
	gob.Register(Student{})
	gob.Register(time.Time{})
	gob.Register(map[string]string{})

	// Connect to Redis.
	cache := remember.New()

	// Store a simple string in the cache.
	fmt.Println("Putting value in the cache with key of foo")
	err := cache.Set("foo", "bar")
	if err != nil {
		fmt.Println("Error getting foo:", err)
		os.Exit(1)
	}

	fmt.Println("Putting an int value into the cache")
	err = cache.Set("intval", 10)
	if err != nil {
		fmt.Println("Error getting foo:", err)
		os.Exit(1)
	}

	fmt.Println("Pulling values out")
	s, err := cache.Get("foo")
	if err != nil {
		fmt.Println("Error getting foo:", err)
		os.Exit(1)
	}

	i, err := cache.GetInt("intval")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("foo: %s, intval: %d\n", s, i)

	// Create an object to store in the cache.
	mary := Student{Name: "Mary", Age: 10}

	// Put the value into the cache with the key student_mary.
	fmt.Println("Putting student_mary into the cache...")
	err = cache.Set("student_mary", mary)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Pull the value out of the cache.
	fmt.Println("Pulling student_mary from the cache....")
	fromCache, err := cache.Get("student_mary")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	student := fromCache.(Student)

	fmt.Printf("%s is %d years old\n", student.Name, student.Age)

	fmt.Println("student_mary is in cache:", cache.Has("student_mary"))
	fmt.Println("Deleting student_mary from the cache....")
	_ = cache.Forget("student_mary")
	fmt.Println("student_mary is in cache after delete:", cache.Has("student_mary"))

	now := time.Now()
	
	// Put a time.Time type in the cache.
	err = cache.Set("now", now)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Retrieve time.Time from the cache.
	n, err := cache.GetTime("now")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(n.Format("2006-01-02 03:04:05 PM"))

	// Create a map.
	myMap := make(map[string]string)
	myMap["1"] = "A"
	myMap["2"] = "B"
	myMap["3"] = "C"

	// Put the map in the cache.
	err = cache.Set("mymap", myMap)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Pull the map out of the cache.
	m, err := cache.Get("mymap")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	myMap2 := m.(map[string]string)
	fmt.Println("1 is", myMap2["1"])

	fmt.Println("Setting 3 foo vars....")
	cache.Set("fooa", "bar")
	cache.Set("foob", "bar")
	cache.Set("fooc", "bar")

	// Remove all entries from the cache with the prefix "foo".
	err = cache.EmptyByMatch("foo")
	if err != nil {
		log.Println(err)
	}

	fmt.Println("cache has fooa:", cache.Has("fooa"))
}
~~~