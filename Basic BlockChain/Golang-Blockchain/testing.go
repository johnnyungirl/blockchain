package main

import (
	"fmt"
)

func birthdayCakeCandles(candles []int32) int32 {
    count:=0
    max:=candles[0]
    for i:=0;i<len(candles);i++{
        if(candles[i]>max){
            max=candles[i]
            count=0
        }
        if(candles[i]==max){
            count++
        }
    }
    // Write your code here
    return int32(count)
}

func main() {
	 arr:= []int32{56, 221 ,2,8974,1,2,3,8974}
	fmt.Println(birthdayCakeCandles(arr))
}
