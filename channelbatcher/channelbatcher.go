package channelbatcher

import "github.com/frontierpsycho/paradoxutil/log"

var logger = log.CreateLog("channelbatcher")

func BatchChannel(channel chan string, batchedChannel chan []string, batchSize int) {
	buff := make([]string, 0, batchSize)

	for {
		next, more := <-channel
		if more {
			buff = append(buff, next)

			if len(buff) >= batchSize {
				batchedChannel <- buff
				buff = make([]string, 0, batchSize)
			}
		} else {
			if len(buff) > 0 {
				batchedChannel <- buff
			}
			close(batchedChannel)
			return
		}
	}
}
