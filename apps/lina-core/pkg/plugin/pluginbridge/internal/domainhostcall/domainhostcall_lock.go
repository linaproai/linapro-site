// This file implements the guest-side distributed lock host-service client.

package domainhostcall

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// lockService adapts the lock host service to lockcap.Service.
type lockService struct{ baseService }

// Lock creates the distributed lock domain guest client.
func Lock(invoker HostServiceInvoker) lockcap.Service {
	return &lockService{baseService: newBaseServiceWithHostService(nil, invoker)}
}

// Acquire attempts to acquire one governed distributed lock.
func (s *lockService) Acquire(_ context.Context, in lockcap.AcquireInput) (*lockcap.AcquireOutput, error) {
	response := &protocol.HostServiceLockAcquireResponse{}
	err := s.callHostServiceJSONRequest(
		protocol.HostServiceLock,
		protocol.HostServiceMethodLockAcquire,
		in.Name,
		"",
		protocol.HostServiceLockAcquireRequest{LeaseMillis: leaseMillis(in.Lease)},
		response,
	)
	if err != nil {
		return nil, err
	}
	return &lockcap.AcquireOutput{
		Acquired: response.Acquired,
		Ticket:   response.Ticket,
		ExpireAt: parseWireTime(response.ExpireAt),
	}, nil
}

// Renew extends one governed distributed lock using the issued ticket.
func (s *lockService) Renew(_ context.Context, in lockcap.RenewInput) (*lockcap.RenewOutput, error) {
	response := &protocol.HostServiceLockRenewResponse{}
	err := s.callHostServiceJSONRequest(
		protocol.HostServiceLock,
		protocol.HostServiceMethodLockRenew,
		in.Name,
		"",
		protocol.HostServiceLockRenewRequest{Ticket: in.Ticket},
		response,
	)
	if err != nil {
		return nil, err
	}
	return &lockcap.RenewOutput{ExpireAt: parseWireTime(response.ExpireAt)}, nil
}

// Release releases one governed distributed lock using the issued ticket.
func (s *lockService) Release(_ context.Context, in lockcap.ReleaseInput) error {
	return s.callHostServiceJSONRequest(
		protocol.HostServiceLock,
		protocol.HostServiceMethodLockRelease,
		in.Name,
		"",
		protocol.HostServiceLockReleaseRequest{Ticket: in.Ticket},
		nil,
	)
}

func leaseMillis(lease time.Duration) int64 {
	if lease <= 0 {
		return 0
	}
	return int64(lease / time.Millisecond)
}
