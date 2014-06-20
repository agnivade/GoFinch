Introduction
============
The implementation of the Finch API in Go.

What is this Finch anyway ?
---------------------------

Finch is a cute robot from Carnegie Mellon's CREATE lab. It is connected simply through a USB cable to a computer and APIs to communicate with the Finch exist in a lot of languages.

Please refer to the [Finch Home Page](http://www.finchrobot.com/) for further details.

Installation
============

Right now, this is only tested on Ubuntu 12.04 machine.

Assumptions
-----------
1. Your code workspace must be structured like mentioned here - http://golang.org/doc/code.html

Pre requisites
--------------
- You need to have [Go](http://golang.org/) installed on your system.

- Install libusb.
```
sudo apt-get install libusb-1.0-0-dev
```
- You need to have libhidapi installed in your system.
```
git clone git@github.com:signal11/hidapi.git
cd hidapi
./bootstrap
./configure

cd libusb
make
sudo make install

cd /usr/local/lib
sudo ln -s libhidapi-libusb.so libhidapi.so
```
On my system, it looks like this-
```
ls -l /usr/local/lib/ | grep libusb
-rw-r--r-- 1 root root   77918 Jun 12 01:18 libhidapi-libusb.a
-rwxr-xr-x 1 root root    1017 Jun 12 01:18 libhidapi-libusb.la
lrwxrwxrwx 1 root root      25 Jun 12 01:18 libhidapi-libusb.so -> libhidapi-libusb.so.0.0.0
lrwxrwxrwx 1 root root      25 Jun 12 01:18 libhidapi-libusb.so.0 -> libhidapi-libusb.so.0.0.0
-rwxr-xr-x 1 root root   67598 Jun 12 01:18 libhidapi-libusb.so.0.0.0
lrwxrwxrwx 1 root root      21 Jun 12 01:57 libhidapi.so -> libhidapi-libusb.so.0
```
- You need to have a Finch to be able to use the API. (Well duh..)
- Please set up your [GOPATH](http://golang.org/doc/code.html#GOPATH) properly.

Now you need to download this package. Enter this inside your "src" folder.
```
go get github.com/agnivade/GoFinch
```
Now start using the API and have fun !

Documentation
=============

Documentation for this site can be found here- https://godoc.org/github.com/agnivade/GoFinch

Examples
========

You might need to set the LD_LIBRARY_PATH environment variable to /usr/local/lib directory before running the main program.
```go
package main

import (
  "log"
  "time"
  "github.com/agnivade/GoFinch"
)

func main() {

  finch_handle, err := finch.Open()
  if err != nil {
    log.Fatalf("Could not connect to Finch: %s", err)
    return
  }
  defer finch_handle.Close()

  _, err = finch_handle.SetBuzzer(1000, 800, true)
  if err != nil {
    log.Fatalf("Error in setting Finch Buzzer: %s", err)
    return
  }

  var temp float64
  temp, err = finch_handle.GetTemperature()
  if err != nil {
    log.Fatalf("Error in getting Finch temperature: %s", err)
    return
  }
  log.Printf("The temperature is- %f", temp)

  var left_sensor, right_sensor byte
  left_sensor, right_sensor, err = finch_handle.GetLight()
  if err != nil {
    log.Fatalf("Error in getting Finch Light: %s", err)
    return
  }
  log.Printf("The light is- Left: %d, Right: %d", left_sensor, right_sensor)

  _, err = finch_handle.SetLed(100,200,0)
  if err != nil {
    log.Fatalf("Error in setting Finch LED- %s", err)
    return
  }
  time.Sleep(2*time.Second)
}
```


