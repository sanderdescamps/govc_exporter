package memory_db

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type Table struct {
	data            map[string]*Item
	lock            sync.RWMutex
	cleanupInterval time.Duration
	stopChan        chan struct{}
}

func NewTable() *Table {
	return &Table{
		data: make(map[string]*Item),
	}
}
func (t *Table) Set(key string, value interface{}) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.data[key] = NewItem(value)
	return nil
}

func (t *Table) SetWithTTL(key string, value interface{}, TTL time.Duration) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.data[key] = NewItemWithTTL(value, TTL)
	return nil
}

func (t *Table) Get(key string, res interface{}) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	val, exists := t.data[key]
	if !exists {
		return errors.New("key not found")
	}

	resv := reflect.ValueOf(res)
	if resv.Kind() != reflect.Pointer || resv.IsNil() {
		return errors.New("input error. Value must be a pointer")
	}

	valv := reflect.ValueOf(val)
	if !valv.Type().AssignableTo(resv.Elem().Type()) {
		return fmt.Errorf("type mismatch: cannot assign %s to %s",
			valv.Type(), resv.Elem().Type())
	}

	resv.Elem().Set(valv)
	return nil
}

func (t *Table) FindByProp(propName string, propValue interface{}, res interface{}) error {
	outVal := reflect.ValueOf(res)
	if outVal.Kind() != reflect.Ptr || outVal.Elem().Kind() != reflect.Slice {
		return errors.New("input error. Value 'res' must be a pointer to a slice")
	}

	sliceVal := outVal.Elem()

	for _, val := range t.data {
		rv := reflect.ValueOf(val)

		// Dereference pointer if needed
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Struct {
			continue
		}

		field := rv.FieldByName(propName)
		if field.IsValid() && field.CanInterface() && reflect.DeepEqual(field.Interface(), propValue) {
			// Add the original value (may be struct or pointer) to the output slice
			sliceVal.Set(reflect.Append(sliceVal, reflect.ValueOf(val)))
		}
	}
	return nil
}

func (t *Table) GetAll(res interface{}) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	resv := reflect.ValueOf(res)
	if (resv.Kind() != reflect.Ptr && resv.Elem().Kind() == reflect.Slice) || resv.IsNil() {
		return errors.New("output must be a non-nil pointer to a slice")
	}

	slicev := resv.Elem()
	if slicev.Kind() != reflect.Slice {
		return errors.New("output must be a pointer to a slice")
	}

	elemType := slicev.Type().Elem()

	for _, i := range t.data {
		if !i.Expired() {
			val := reflect.ValueOf(i.value)
			if !val.Type().AssignableTo(elemType) {
				return fmt.Errorf("type mismatch: cannot assign %s to slice of %s", val.Type(), elemType)
			}

			slicev.Set(reflect.Append(slicev, val))
		}
	}

	return nil
}

func (s *Table) StartCleaner() {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				func() {
					s.lock.Lock()
					defer s.lock.Unlock()
					for k, v := range s.data {
						if v.Expired() {
							delete(s.data, k)
						}
					}
				}()
			case <-s.stopChan:
				return
			}
		}
	}()
}

func (s *Table) StopCleaner() {
	close(s.stopChan)
}
