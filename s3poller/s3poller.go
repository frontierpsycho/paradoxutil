package s3poller

import (
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"github.com/frontierpsycho/paradoxutil/log"
	"time"
)

var logger *log.LogRepo = log.CreateLog("s3poller")

type S3Poller struct {
	Auth           *aws.Auth
	S3Client       *s3.S3
	Bucket         *s3.Bucket
	EnabledPrefix  string
	Seen           map[string]string
	HandleAddition func([]byte, s3.Key) error
	HandleRemoval  func(string) error
}

func (w *S3Poller) Poll() {

	for {
		more := true
		marker := ""
		newSeen := make([]string, 0)
		for more {
			more = false

			resp, err := w.Bucket.List(w.EnabledPrefix, "", marker, 1000)
			if err != nil {
				logger.Error("%q", err)
				continue
			}
			for _, content := range resp.Contents {

				newSeen = append(newSeen, content.Key)
				if m, exist := w.Seen[content.Key]; exist && m == content.LastModified {
					continue
				}
				w.Seen[content.Key] = content.LastModified
				data, err := w.Bucket.Get(content.Key)
				if err != nil {
					logger.Error("%q", err)
					continue
				}

				err = w.HandleAddition(data, content)
				if err != nil {
					logger.Error("%q", err)
					continue
				}
			}
			more = resp.IsTruncated
			marker = resp.Marker
		}

		for key, _ := range w.Seen {
			found := false
			for _, key2 := range newSeen {
				if key == key2 {
					found = true
					break
				}
			}
			if !found {
				err := w.HandleRemoval(key)
				if err != nil {
					logger.Error("%q", err)
					continue
				}
				delete(w.Seen, key)
			}
		}

		time.Sleep(5 * time.Second)
	}
}
