// this calculate `viscosity`
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sort"
	"sync"
)

var base = 10

type seq []int
type resultSet map[int]int

func (s seq) toInt() (result int) {
	for _, k := range s {
		result = result*base + k
	}
	return
}

// sorted
var prefix = []seq{
	seq{}, seq{2}, seq{3}, seq{4}, seq{6}, seq{7}, seq{8}, seq{9}, seq{2, 6}, seq{3, 5},
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get max mv from command argument
	max, _ := strconv.Atoi(os.Args[1])

	// run and get result
	result := Run(ctx, max)
	// well-known values
	result[0] = 0
	result[1] = 10
	result[2] = 25

	fmt.Println("Results:")
	// sort by key
	keys := []int{}
	for k, _ := range result {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Printf("%d,%d\n", k, result[k])
	}
}

func Run(ctx context.Context, max int) (result resultSet) {
	return checkController(ctx, max)
}

// Control flow
func checkController(ctx context.Context, max int) (result resultSet) {
	// result slice
	result = map[int]int{}

	// create go-routine
	// Prefix ごとに計算することにする
	numOfGoRoutine := 0
	var mu sync.Mutex
	// 各prefix のgoroutineにseqを送る
	mapDataChan := map[int](chan seq){}
	// result が更新されたら送る
	mapToPrefixChan := map[int](chan ([2]int)){}
	// result よりも良い結果が出たときにcontrol flowへ連絡する
	fromPrefixChan := make(chan ([2]int), 1)
	// prefix goroutine が終わったときに連絡してもらう
	commonDoneChan := make(chan seq, 20)
	for _, p := range prefix {
		numOfGoRoutine++
		func() {
			dataChan := make(chan seq, 1)
			mapDataChan[p.toInt()] = dataChan
			toPrefixChan := make(chan [2]int, 1)
			mapToPrefixChan[p.toInt()] = toPrefixChan
			// run
			go checkWithPrefix(ctx, p, &mu, commonDoneChan, dataChan, toPrefixChan, fromPrefixChan, max)
		}()
	}

	// generate formal seq
	generationChan := make(chan seq, 1)
	doneGeneration := make(chan struct{})
	var length = 1
	go generateFormalSequenceWithLength(ctx, doneGeneration, generationChan, length)

	// Control Flow
	for {
		select {
		case v := <-generationChan:
			for _, ch := range mapDataChan {
				go func(ch chan seq, v seq) {
					ch <- v
				}(ch, v)
			}
		case v := <-fromPrefixChan:
			//fmt.Println("fromPrefixChan", v)
			degree, number := v[0], v[1]
			mu.Lock()

			if (result[degree] == 0) || (result[degree] > number) {
				result[degree] = number
			}
			mu.Unlock()
			for _, ch := range mapToPrefixChan {
				//fmt.Println("mapto prefix chan", k)
				go func(ch chan [2]int, v [2]int) {
					ch <- v
				}(ch, v)
			}
		case v := <-commonDoneChan:
			//fmt.Println("commonDone", v)
			mu.Lock()
			delete(mapDataChan, v.toInt())
			delete(mapToPrefixChan, v.toInt())
			mu.Unlock()
			numOfGoRoutine--
			//fmt.Println("Rest of goroutine: ", numOfGoRoutine)
			if numOfGoRoutine == 0 {
				return result
			}
		case _ = <-doneGeneration:
			//fmt.Println("doneGeneration")
			length++
			go generateFormalSequenceWithLength(ctx, doneGeneration, generationChan, length)
		}
	}
	return
}

func checkWithPrefix(ctx context.Context, p seq, mu *sync.Mutex, doneChan chan seq,
	dataChan chan seq, toPrefix chan [2]int, fromPrefix chan [2]int, max int, ) {
	var i int
	var currentResult = resultSet{}
	var done = false
	var tmpDone bool
	var isNew bool
	v := V{base: base}
	for {
		select {
		case d := <-toPrefix:
			currentResult[d[0]] = d[1]
		case s := <-dataChan:
			//fmt.Println("datachan at prefix ", p, s)
			// append prefix
			if len(p) != 0 {
				s = append(p, s...)
			}

			number := s.toInt()
			// もし、自分が今やっているものが既知の結果よりも大きい数字の場合は終わる
			// ただし i>=3 だけ考える
			for i = 3; i < max; i++ {
				tmpDone = (currentResult[i] != 0) && (currentResult[i] < number)
				if !tmpDone {
					done = false
					break
				} else {
					done = true
				}
			}
			if done {
				//fmt.Println("done prefix", p, number, currentResult)
				doneChan <- p
				return
			}
			degree := v.calcMV(number)
			mu.Lock()
			// 値を更新できそうかチェックする
			isNew = (degree < max) && ((currentResult[degree] == 0) || (currentResult[degree] > number))
			mu.Unlock()
			if (isNew) {
				// if updated, send data to controller
				go func() {
					fromPrefix <- [2]int{degree, number}
				}()
			}
		}
	}
}

func generateFormalSequenceWithLength(ctx context.Context, doneGeneration chan struct{}, ch chan seq, length int) {
	var choice = seq{7, 8, 9}

	gen(ctx, ch, length, choice)
	select {
	case doneGeneration <- struct{}{}:
		// do nothing
		return
	case <-ctx.Done():
		return
	}
}

func gen(ctx context.Context, ch chan seq, length int, choice seq) {
	var a, b, c int
	for a = length; a >= 0; a-- {
		for b = length - a; b >= 0; b-- {
			c = length - a - b
			select {
			case <-ctx.Done():
				return
			case ch <- getSeq(choice, a, b, c):
				// do nothing
			}
		}
	}
	return
}

func getSeq(choice seq, a, b, c int) (result seq) {
	result = []int{}
	for i := 0; i < a; i++ {
		result = append(result, choice[0])
	}
	for i := 0; i < b; i++ {
		result = append(result, choice[1])
	}
	for i := 0; i < c; i++ {
		result = append(result, choice[2])
	}
	return
}
