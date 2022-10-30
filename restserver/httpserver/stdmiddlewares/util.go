package stdmiddlewares

import "time"

func getNowUnix() int64 {
	return time.Now().Unix()
}

func sinceUnix(ts int64) int64 {
	now := getNowUnix()
	return now - ts
}
