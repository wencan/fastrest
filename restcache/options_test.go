package restcache

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/wencan/fastrest/restcache/mock_restcache"
)

type testResponseWithValid struct {
	Valid bool
}

func (resp testResponseWithValid) IsValidCache() bool {
	return resp.Valid
}

func TestGetWithValid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	mockStorage := mock_restcache.NewMockStorage(ctrl)
	mockStorage.EXPECT().Get(gomock.Eq(ctx), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf((*testResponseWithValid)(nil))).DoAndReturn(func(ctx context.Context, key string, destPtr interface{}) (found bool, err error) {
		if key == "invalid" {
			reflect.ValueOf(destPtr).Elem().Set(reflect.ValueOf(testResponseWithValid{Valid: false}))
		} else {
			reflect.ValueOf(destPtr).Elem().Set(reflect.ValueOf(testResponseWithValid{Valid: true}))
		}
		return true, nil
	}).AnyTimes()

	caching := Caching{
		Storage:  mockStorage,
		Query:    func(ctx context.Context, destPtr, args interface{}) (found bool, err error) { return false, nil },
		TTLRange: [2]time.Duration{time.Minute * 4, time.Minute * 6},
	}
	var response testResponseWithValid
	found, err := caching.Get(ctx, &response, "valid", nil)
	if assert.Nil(t, err) {
		assert.True(t, found)
	}
	found, err = caching.Get(ctx, &response, "invalid", nil)
	if assert.Nil(t, err) {
		assert.False(t, found)
	}
}

func TestMGetWithValid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock_restcache.NewMockMStorage(ctrl)
	mockStorage.EXPECT().MGet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any()).DoAndReturn(func(ctx context.Context, keys []string, destSlicePtr interface{}) (missIndexes []int, err error) {
		var destSliceValue = reflect.ValueOf(destSlicePtr).Elem()
		for index, key := range keys {
			if strings.HasPrefix(key, "miss") {
				missIndexes = append(missIndexes, index)
			} else if strings.HasPrefix(key, "invalid") {
				destSliceValue.Set(reflect.Append(destSliceValue, reflect.ValueOf(testResponseWithValid{false})))
			} else {
				destSliceValue.Set(reflect.Append(destSliceValue, reflect.ValueOf(testResponseWithValid{true})))
			}
		}
		return missIndexes, nil
	}).AnyTimes()
	mockStorage.EXPECT().MSet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any(), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, keys []string, destSlice interface{}, ttl time.Duration) error {
		return nil
	}).AnyTimes()

	mquery := func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
		reqs := argsSlice.([]string)
		for index := range reqs {
			missIndexes = append(missIndexes, index)
		}
		return missIndexes, nil
	}

	mcaching := MCaching{
		MStorage:    mockStorage,
		MQuery:      mquery,
		TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
		SentinelTTL: time.Second,
	}

	keys := []string{"miss_0", "valid_1", "invalid_2", "miss_3", "valid_4", "invalid_5"}
	args := keys
	var results []testResponseWithValid
	missIndexes, err := mcaching.MGet(context.TODO(), &results, keys, args)
	if assert.Nil(t, err) {
		assert.Equal(t, []int{0, 2, 3, 5}, missIndexes)
	}
}
