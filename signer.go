package main

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := sync.WaitGroup{}
	in := make(chan interface{})

	for _, jobFunc := range jobs {
		out := make(chan interface{})
		wg.Add(1)
		go func(j job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			j(in, out)
		}(jobFunc, in, out)
		in = out
	}

	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for data := range in {
		wg.Add(1)
		go func(d interface{}) {
			defer wg.Done()
			dataInt, ok := d.(int)
			if !ok {
				log.Println("Ожидается int")
			}
			dataString := strconv.Itoa(dataInt)

			mu.Lock()
			md5Hash := DataSignerMd5(dataString)
			mu.Unlock()

			crc32Chan := make(chan string, 1)
			crc32md5Chan := make(chan string, 1)

			go func() {
				crc32Chan <- DataSignerCrc32(dataString)
			}()

			go func() {
				crc32md5Chan <- DataSignerCrc32(md5Hash)
			}()

			out <- <-crc32Chan + "~" + <-crc32md5Chan
		}(data)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	quantityAction := 6
	wg := sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		go func(d interface{}) {
			defer wg.Done()
			dataStr, ok := d.(string)
			if !ok {
				log.Println("Ожидается string")
			}

			results := make([]string, quantityAction)
			innerWg := sync.WaitGroup{}

			for i := 0; i < quantityAction; i++ {
				innerWg.Add(1)
				go func(idx int) {
					defer innerWg.Done()
					results[idx] = DataSignerCrc32(strconv.Itoa(idx) + dataStr)
				}(i)
			}

			innerWg.Wait()
			out <- strings.Join(results, "")
		}(data)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string

	for data := range in {
		res, ok := data.(string)
		if !ok {
			log.Println("Ожидается string")
		}
		results = append(results, res)
	}

	sort.Strings(results)
	out <- strings.Join(results, "_")
}
