package log

import "context"

func GetLogWritter(name string) *LogWritter {
	return &LogWritter{log: GetLogger(name)}
}

type LogWritter struct {
	log *Log
}

func (w *LogWritter) Write(p []byte) (n int, err error) {
	w.log.Debugw(context.Background(), string(p[:]))
	return len(p), nil
}
