package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

const TimeFormat = "20060102150405.000"

func main() {
	var wg sync.WaitGroup
	var chans = []chan string{
		make(chan string),
		make(chan string),
		make(chan string),
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT)
	go func(chans []chan string) {
		<-sigc
		fmt.Println("Got a control C event")
		for _, c := range chans {
			fmt.Println("Sending MSG to", c)
			c <- "I should probabyl die now!!!"
		}
	}(chans)
	s := session.New(&aws.Config{Region: aws.String("us-west-2")})
	kc := kinesis.New(s)
  streamName := "cm_message_kinesis_kinesis_stream"

	  wg.Add(3)
    go write_kinesis(*kc, &streamName, &wg, &chans[0])
    go write_kinesis(*kc, &streamName, &wg, &chans[1])
    go write_kinesis(*kc, &streamName, &wg, &chans[2])

//	go write_vomit("./stream1.dat", &wg, &chans[0])
//	go write_vomit("./stream2.dat", &wg, &chans[1])
//	go write_vomit("./stream3.dat", &wg, &chans[2])
	wg.Wait()
}

func write_kinesis(kc kinesis.Kinesis, streamName *string, wg *sync.WaitGroup, done *chan string) {
	titles := NewTitleCollection(1000)

  i := 0
	finished := false
	for {
		select {
		case <-*done:
			finished = true
		default:
			if finished {
				fmt.Println("done writing")
				wg.Done()
				return
			}
      time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
      //write record
      title  := titles.Pop()
		  fmt.Println(title)
			_, err := kc.PutRecord(&kinesis.PutRecordInput{
        Data:         []byte(title),
        StreamName:   streamName,
        PartitionKey: aws.String("None"),
      })
      if err != nil {
        fmt.Println("Could not write to kinesis stream")
      }
      i++
		}
	}
	wg.Done()
}

func write_vomit(fname string, wg *sync.WaitGroup, done *chan string) {
	titles := NewTitleCollection(1000)
	f, err := os.OpenFile(fname, syscall.O_APPEND|syscall.O_CREAT|syscall.O_WRONLY, 0644)
	defer f.Close()
	defer close(*done)
	if err != nil {
		fmt.Println("Could not open file for writing")
	}
	w := csv.NewWriter(f)
	i := 0
	finished := false
	for {
		select {
		case <-*done:
			finished = true
		default:
			if finished {
				fmt.Println("no more writing for", fname)
				w.Flush()
				wg.Done()
				return
			}
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			if err := w.Write(gen_record(i, fname, titles.Pop())); err != nil {
				fmt.Println("could not write to csv")
			}
			i++
		}
	}

	if err := w.Error(); err != nil {
		fmt.Println(err)
		fmt.Println("boom")
	}
	wg.Done()
}

func gen_record(seq int, fname string, title string) []string {
	rec := []string{
		the_time(),
		fname,
		strconv.Itoa(seq),
		title,
	}
	return rec
}

func the_time() string {
	t := time.Now()
	s := t.Format(TimeFormat)
	return s
}
