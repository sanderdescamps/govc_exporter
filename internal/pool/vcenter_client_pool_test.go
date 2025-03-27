package pool_test

import (
	"context"
	"testing"

	"github.com/sanderdescamps/govc_exporter/internal/pool"
)

func Test_VCenterClientPool(t *testing.T) {
	ctx := context.Background()
	pool, err := pool.NewVCenterClientPoolWithLogger(
		ctx,
		"https://localhost:8989",
		"testuser",
		"testpass",
		5,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create pool %v", err)
	}

	client, release, err := pool.Acquire()
	if err != nil {
		t.Errorf(err.Error())
	} else {
		t.Logf("client valid: %t", client.Valid())
		release()
	}

	client, release, err = pool.Acquire()
	if err != nil {
		t.Errorf(err.Error())
	} else {
		t.Logf("client valid: %t", client.Valid())
		release()
	}

	client, release, err = pool.Acquire()
	if err != nil {
		t.Errorf(err.Error())
	} else {
		t.Logf("client valid: %t", client.Valid())
		release()
	}

}
