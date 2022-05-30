/*
Copyright AppsCode Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/madflojo/tasks"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/syndtr/goleveldb/leveldb"
)

var buffers = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

type Scheduler struct {
	db *leveldb.DB
	s  *tasks.Scheduler
}

func NewScheduler() (*Scheduler, error) {
	path, _ := ioutil.TempDir("", "tasks")
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &Scheduler{
		s:  tasks.New(),
		db: db,
	}, nil
}

func (s *Scheduler) Close() error {
	s.s.Stop()
	return s.db.Close()
}

func (s *Scheduler) Cleanup(fn func([]byte) error) error {
	iter := s.db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := iter.Key()
		val := iter.Value()

		del := false
		ts, args, ok := bytes.Cut(val, []byte("|"))
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "key %s has invalid data %s", string(key), string(val))
			del = true
		} else {
			t, err := time.Parse(time.RFC3339, string(ts))
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "key %s has invalid end timestamp %s, err: %v", string(key), string(ts), err)
				del = true
			} else {
				if time.Now().After(t) {
					_, _ = fmt.Fprintf(os.Stdout, "cleaning up task id %s with args %s", string(key), string(args))
					if err := fn(args); err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "failed to execute cleanup function for task id %s, args %s, err: %v", string(key), string(args), err)
					}
					del = true
				} else {
					if task, _ := s.s.Lookup(string(key)); task == nil {
						// re-schedule
						interval := time.Until(t)
						err := s.s.AddWithID(string(key), &tasks.Task{
							Interval: interval,
							RunOnce:  true,
							TaskFunc: func() error {
								if err := fn(args); err != nil {
									return err
								}
								return s.db.Delete(key, nil)
							},
							ErrFunc: func(e error) {
								_ = s.db.Delete(key, nil)
								_, _ = fmt.Fprintf(os.Stderr, "an error occurred when executing task %s - %v", string(key), e)
							},
						})
						if err == tasks.ErrIDInUse {
							continue
						} else if err != nil {
							return errors.Wrapf(err, "failed to schedule task")
						}
					}
				}
			}
		}

		if del {
			if err := s.db.Delete(key, nil); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "failed to delete key %s, args %s, err: %v", string(key), string(args), err)
			}
		}
	}
	iter.Release()
	return iter.Error()
}

func (s *Scheduler) Schedule(t time.Time, fn func([]byte) error, args []byte) error {
	buf := buffers.Get().(*bytes.Buffer)
	defer buffers.Put(buf)

	for {
		id := xid.New()
		interval := time.Until(t)
		err := s.s.AddWithID(id.String(), &tasks.Task{
			Interval: interval,
			RunOnce:  true,
			TaskFunc: func() error {
				if err := fn(args); err != nil {
					return err
				}
				return s.db.Delete(id.Bytes(), nil)
			},
			ErrFunc: func(e error) {
				_ = s.db.Delete(id.Bytes(), nil)
				_, _ = fmt.Fprintf(os.Stderr, "an error occurred when executing task %s - %v", id, e)
			},
		})
		if err == tasks.ErrIDInUse {
			continue
		} else if err != nil {
			return errors.Wrapf(err, "failed to schedule task")
		}

		buf.Reset()
		buf.WriteString(t.Format(time.RFC3339))
		buf.WriteRune('|')
		buf.Write(args)
		if err = s.db.Put(id.Bytes(), buf.Bytes(), nil); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(os.Stdout, "Task with args %s will execute in %s", string(args), interval)
		return nil
	}
}
