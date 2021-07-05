package danmep

import (
	danmipam "github.com/nokia/danm/pkg/ipam"
)

type danmReleaseIPServiceImpl releaseIPServiceImplBase

func (h *danmReleaseIPServiceImpl) IsIPAllocatedByMe(ip string, neType string) bool {
	return danmipam.WasIpAllocatedByDanm(ip, h.dnet.Spec.Options.Cidr) ||
		danmipam.WasIpAllocatedByDanm(ip, h.dnet.Spec.Options.Pool6.Cidr)
}

func (h *danmReleaseIPServiceImpl) ReleaseIP(ip string) error {
	return danmipam.Free(h.danmClient, *h.dnet, ip)
}
