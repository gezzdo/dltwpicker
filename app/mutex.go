package main

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/appengine/memcache"
)

func lock(ctx context.Context, key string) error {
	for tries := 0; tries < 10; tries++ {
		v, err := memcache.Increment(ctx, key, 1, 0)
		if err != nil {
			memcache.Delete(ctx, key)
			return errors.Wrap(err, "failed to lock")
		}
		if v == 1 {
			return nil
		}
		_, err = memcache.IncrementExisting(ctx, key, -1)
		if err != nil {
			memcache.Delete(ctx, key)
			return errors.Wrap(err, "failed to lock")
		}
		time.Sleep(400 * time.Millisecond)
	}
	return errors.New("mutex is busy")
}

func unlock(ctx context.Context, key string) error {
	_, err := memcache.IncrementExisting(ctx, key, -1)
	if err != nil {
		memcache.Delete(ctx, key)
		return errors.Wrap(err, "failed to unlock")
	}
	return nil
}

func transaction(ctx context.Context, key string, f func()) (err error) {
	err = lock(ctx, key)
	if err != nil {
		return err
	}
	defer func() {
		err = unlock(ctx, key)
	}()
	f()
	return
}
