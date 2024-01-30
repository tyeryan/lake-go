package config

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"sync"

	"github.com/google/wire"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const (
	DefStructTagName         = "configstruct"
	defStructTagDefaultValue = "configdefault"
	defTimeLayout            = "200601021504"
)

var (
	WireSet = wire.NewSet(
		ProvideDecodeOption,
		ProvideConfigStoreImpl,
	)
	WireSetWithJsonStringSupport = wire.NewSet(
		ProvideDecodeOptionWithJsonStringSupport,
		ProvideConfigStoreImpl,
	)

	mutex    = &sync.Mutex{}
	instance *ConfigStoreImpl
)

func init() {
	viper.AutomaticEnv()
}

// ConfigStoreImpl config store
type ConfigStoreImpl struct {
	decodeOption viper.DecoderConfigOption
}

// ConfigStore config store interface
type ConfigStore interface {
	GetConfig(val interface{}) error
	SetDefault(key string, val interface{})
}

func getDefaultDecodeHooks() []mapstructure.DecodeHookFunc {
	return []mapstructure.DecodeHookFunc{
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		// adding string to time decoder for which isn't enabled by default in ciper
		// ref: https://github.com/spf13/viper/blob/master/viper.go:838
		// YYYYmmDDHHMM
		mapstructure.StringToTimeHookFunc(defTimeLayout)}
}

// ProvideDecodeOption  decode option provider
func ProvideDecodeOption(ctx context.Context) viper.DecoderConfigOption {
	return func(option *mapstructure.DecoderConfig) {
		option.TagName = DefStructTagName
		option.WeaklyTypedInput = true
		option.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			getDefaultDecodeHooks()...,
		)
	}
}

// Add support to Json String -> struct ptr decode hook
func ProvideDecodeOptionWithJsonStringSupport(ctx context.Context) viper.DecoderConfigOption {
	var newDecodeHookList []mapstructure.DecodeHookFunc
	newDecodeHookList = append(newDecodeHookList, getDefaultDecodeHooks()...)
	newDecodeHookList = append(newDecodeHookList, JsonStringToStructPtrHookFunc())
	return func(option *mapstructure.DecoderConfig) {
		option.TagName = DefStructTagName
		option.WeaklyTypedInput = true
		option.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			newDecodeHookList...,
		)
	}
}

// JsonStringToStructPtrHookFunc returns a DecodeHookFunc that converts
// Json strings to Struct Ptr.
func JsonStringToStructPtrHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t.Kind() != reflect.Ptr {
			return data, nil
		}
		s := strings.Trim(data.(string), " ")
		if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
			return data, nil
		}

		v := reflect.Value{}
		v = reflect.New(t.Elem())
		newP := v.Interface()
		if err := json.Unmarshal([]byte(s), newP); err != nil {
			return nil, err
		}
		return newP, nil
	}
}

// ProvideConfigStoreImpl config store provider
func ProvideConfigStoreImpl(ctx context.Context, option viper.DecoderConfigOption) ConfigStore {
	mutex.Lock()
	defer mutex.Unlock()

	if instance != nil {
		return instance
	}

	instance = &ConfigStoreImpl{
		decodeOption: option,
	}

	return instance
}

// GetConfig get config from env, default val
func (c *ConfigStoreImpl) GetConfig(val interface{}) error {
	c.loadTagConfig(val)
	return viper.Unmarshal(val, c.decodeOption)
}

// SetDefault SetDefault sets the default value for this key.
func (c *ConfigStoreImpl) SetDefault(key string, val interface{}) {
	viper.SetDefault(key, val)
	viper.BindEnv(key)
}

// loadTagConfig load custom tag config
func (c *ConfigStoreImpl) loadTagConfig(val interface{}) error {

	dataValues := reflect.TypeOf(val)
	if dataValues.Kind() != reflect.Ptr {
		return errors.New("only accept struct pointer")
	}
	dataFields := dataValues.Elem()

	for i := 0; i < dataFields.NumField(); i++ {
		field := dataFields.Field(i)

		configKeyTag := field.Tag.Get(DefStructTagName)
		defaultValTag := field.Tag.Get(defStructTagDefaultValue)
		c.SetDefault(configKeyTag, defaultValTag)
	}

	return nil
}
