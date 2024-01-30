package log

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/elliotchance/orderedmap"
	ctxutil "github.com/tyeryan/l-protocol/context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"reflect"
	"strings"
	"time"
)

// Log wraps the underlying logging framework
// and provides default logging of stan (the request id) and logger (who)
type Log struct {
	logger  *zap.SugaredLogger
	infoMap *orderedmap.OrderedMap
	start   time.Time
}

const (
	TimeSpentSeconds = "TimeSpentSeconds"
	RequestId        = "RequestId"
	Error            = "Error"
	Alert            = "Alert"
	Panic            = "Panic"
	ErrorDetail      = "ErrorDetail"
)

var (
	LoggerConfig zapcore.Core
	logLevel     zap.AtomicLevel
)

func init() {
	level := os.Getenv("LOG_LEVEL")
	lvl := zap.InfoLevel // default
	lvl.UnmarshalText([]byte(level))
	logLevel = zap.NewAtomicLevelAt(lvl)

	UseJSONLogger()
}

// UseJSONLogger log as json format, this is the default
func UseJSONLogger() {
	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.LevelKey = "severity" // it seems the "level" is always being overriden in fluentd

	LoggerConfig = zapcore.NewCore(zapcore.NewJSONEncoder(zapConfig), os.Stdout, logLevel)
}

// UseConsoleLogger log as plain text, can be used in development or unit test
func UseConsoleLogger() {
	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	LoggerConfig = zapcore.NewCore(zapcore.NewConsoleEncoder(zapConfig), os.Stdout, logLevel)
}

func EnableDebug() {
	logLevel.SetLevel(zap.DebugLevel)
}

func GetLogger(name string) *Log {
	return &Log{logger: zap.New(LoggerConfig).Sugar().With("logger", string(name)),
		infoMap: orderedmap.NewOrderedMap(), start: time.Now()}
}

func (log *Log) Debugw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, log.infoMap)
	log.logger.Debugw(formattedMsg, concatIndexMessages(formattedIndexMsg, keysAndValues)...)
}

func (log *Log) Infow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, log.infoMap)
	log.logger.Infow(formattedMsg, concatIndexMessages(formattedIndexMsg, keysAndValues)...)
}

func (log *Log) Warnw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, log.infoMap)
	log.logger.Warnw(formattedMsg, concatIndexMessages(formattedIndexMsg, keysAndValues)...)
}

func (log *Log) Warne(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	if err != nil {
		log.infoMap.Set(ErrorDetail, err)
	}

	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, log.infoMap)
	log.logger.Warnw(formattedMsg, concatIndexMessages(formattedIndexMsg, keysAndValues)...)
}

func (log *Log) Errorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, log.infoMap)
	log.logger.Errorw(formattedMsg, concatIndexMessages(formattedIndexMsg, keysAndValues)...)
}

func (log *Log) Errore(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	infoMaps := []*orderedmap.OrderedMap{log.infoMap}
	if err != nil {
		localMap := orderedmap.NewOrderedMap()
		localMap.Set(ErrorDetail, err)
		infoMaps = append(infoMaps, localMap)
	}
	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, infoMaps...)
	log.logger.Errorw(formattedMsg, concatIndexMessages(formattedIndexMsg, keysAndValues)...)
}

func (log *Log) Fatalw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, log.infoMap)
	log.logger.Fatalw(formattedMsg, concatIndexMessages(formattedIndexMsg, keysAndValues)...)
}

func (log *Log) Fatale(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	infoMaps := []*orderedmap.OrderedMap{log.infoMap}
	if err != nil {
		localMap := orderedmap.NewOrderedMap()
		localMap.Set(ErrorDetail, err)
		infoMaps = append(infoMaps, localMap)
	}
	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, infoMaps...)
	log.logger.Fatalw(formattedMsg, concatIndexMessages(formattedIndexMsg, keysAndValues)...)
}

// Alertw will generate alert
func (log *Log) Alertw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	infoMaps := []*orderedmap.OrderedMap{log.infoMap}
	localMap := orderedmap.NewOrderedMap()

	localMap.Set(Alert, true)
	infoMaps = append(infoMaps, localMap)

	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, infoMaps...)
	log.logger.Errorw(formattedMsg, concatIndexMessages(formattedIndexMsg, keysAndValues)...)
}

// Alerte will generate alert
func (log *Log) Alerte(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	infoMaps := []*orderedmap.OrderedMap{log.infoMap}
	localMap := orderedmap.NewOrderedMap()

	localMap.Set(Alert, true)
	if err != nil {
		localMap.Set(ErrorDetail, err)
	}
	infoMaps = append(infoMaps, localMap)

	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, infoMaps...)
	log.logger.Errorw(formattedMsg, concatIndexMessages(formattedIndexMsg, keysAndValues)...)
}

func (log *Log) Canonical(ctx context.Context, msg string, err error, r interface{}) {
	stan, _ := ctxutil.Read(ctx, ctxutil.Stan)
	userID, _ := ctxutil.Read(ctx, ctxutil.UserID)

	infoMaps := []*orderedmap.OrderedMap{log.infoMap}
	localMap := orderedmap.NewOrderedMap()

	localMap.Set(TimeSpentSeconds, time.Since(log.start).Seconds())
	if err != nil {
		localMap.Set(Error, true)
		localMap.Set(ErrorDetail, err)
	}
	if r != nil {
		localMap.Set(Panic, true)
	}
	infoMaps = append(infoMaps, localMap)

	formattedMsg, formattedIndexMsg := formatInfoMap(msg, stan, userID, infoMaps...)
	formattedIndexMsg = insert(formattedIndexMsg, "canonical", true)

	log.logger.Infow(formattedMsg, formattedIndexMsg...)

	if r != nil {
		panic(r)
	}
}

func (log *Log) Add(keysAndValues ...interface{}) {
	if len(keysAndValues) == 0 || len(keysAndValues)%2 != 0 {
		return
	}

	for i := 0; i < len(keysAndValues)-1; i = i + 2 {
		key := strings.ReplaceAll(keysAndValues[i].(string), ".", "")
		log.infoMap.Set(key, keysAndValues[i+1])
	}
}

func insert(slice []interface{}, elements ...interface{}) []interface{} {
	return append(elements, slice...)
}

// value will be formatted as string
// to avoid elasticsearch to marshal/unmarshal as json
func formatKeysAndValues(stan string, userID string, keysAndValues []interface{}) []interface{} {
	things := []interface{}{"stan", stan, "UserID", userID}

	if len(keysAndValues) == 0 {
		return things
	}

	if len(keysAndValues)%2 != 0 {
		return append(things, "everything", fmt.Sprintf("%+v", keysAndValues))
	}

	for i := 0; i < len(keysAndValues)-1; i = i + 2 {
		key := strings.ReplaceAll(keysAndValues[i].(string), ".", "")
		things = append(things, key, spew.Sprintf("%+v", keysAndValues[i+1]))
	}
	return things
}

func concatIndexMessages(existedMessages []interface{}, keysAndValues []interface{}) []interface{} {
	things := existedMessages

	if len(keysAndValues) == 0 {
		return things
	}

	if len(keysAndValues)%2 != 0 {
		return append(things, "everything", fmt.Sprintf("%+v", keysAndValues))
	}

	for i := 0; i < len(keysAndValues)-1; i = i + 2 {
		key := strings.ReplaceAll(keysAndValues[i].(string), ".", "")
		things = append(things, key, spew.Sprintf("%+v", keysAndValues[i+1]))
	}
	return things
}

func formatInfoMap(msg string, stan string, userID string, infoMaps ...*orderedmap.OrderedMap) (string, []interface{}) {
	indexFields := []string{TimeSpentSeconds, RequestId, Error, Panic, Alert}
	things := []interface{}{"stan", stan, "UserID", userID}

	if len(infoMaps) == 0 {
		return msg, things
	}

	processedMsg := msg
	for _, infoMap := range infoMaps {
		if infoMap == nil {
			break
		}

		for _, k := range infoMap.Keys() {
			val, _ := infoMap.Get(k)
			key := strings.ReplaceAll(k.(string), ".", "")

			// append it as index field
			if contains(indexFields, k.(string)) {
				things = append(things, key, val)
			} else {
				// append it in msg
				if "string" == reflect.TypeOf(val).String() {
					val = "'" + spew.Sprintf("%+v", val) + "'"
				}
				processedMsg = processedMsg + " " + key + "=" + spew.Sprintf("%+v", val) + ""
			}
		}
	}

	return processedMsg, things
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
