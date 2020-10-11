package main

import (
	"fmt"
	"time"
)

func main()  {
	const layout = time.RFC850

	t := time.Now()
	fmt.Printf("%s\n",t.Format(layout))
}
