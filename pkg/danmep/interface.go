package danmep

import (
    "fmt"

    danmtypes "github.com/nokia/danm/crd/apis/danm/v1"
    danmclientset "github.com/nokia/danm/crd/client/clientset/versioned"
    "github.com/nokia/danm/pkg/danmep"
    "github.com/nokia/danm/pkg/ipam"
)

// GetSupportedReleaseIPHandlers gets all supported static CNI implementation that cleaner can use
// when deleting dangling DanmEps
func GetSupportedReleaseIPHandlers() []*ReleaseIPService {
    return []*ReleaseIPService{
        NewReleaseIPService(&calicoReleaseIPServiceImpl{}),
    }
}

// ReleaseIPInterface need to be implemented by every static CNI IPAM release plugin
// that want to invoked when Cleaner is cleaning up dangling DanmEps
type ReleaseIPInterface interface {
    ReleaseIP(dnet *danmtypes.DanmNet, ep *danmtypes.DanmEp, ip string) error
    IsIPAllocatedByMe(dnet *danmtypes.DanmNet, ep *danmtypes.DanmEp, ip string) bool
}

type ReleaseIPService struct {
    handler ReleaseIPInterface
}

// NewReleaseIPService constructor encapsulates different ReleaseIPInterface implementations under a generic API
func NewReleaseIPService(handler ReleaseIPInterface) *ReleaseIPService {
    return &ReleaseIPService{handler}
}

// IsIPAllocatedByMe invokes embedded implementation specific IsIPAllocatedByMe function
func (d ReleaseIPService) IsIPAllocatedByMe(dnet *danmtypes.DanmNet, ep *danmtypes.DanmEp, ip string) bool {
    return d.handler.IsIPAllocatedByMe(dnet, ep, ip)
}

// ReleaseIP invokes embedded implementation specific DeleteDanmEp function
func (d ReleaseIPService) ReleaseIP(dnet *danmtypes.DanmNet, ep *danmtypes.DanmEp, ip string) error {
    return d.handler.ReleaseIP( dnet, ep, ip)
}

// SelectAppropriateServiceImplementation looks up and returns the
// first suitable ReleaseIP service implementation for given IP address
// returns nil when no suitable implementation found
func SelectAppropriateServiceImplementation(dnet *danmtypes.DanmNet, ep *danmtypes.DanmEp, ip string) *ReleaseIPService {
    for _, handler := range GetSupportedReleaseIPHandlers() {
        if handler.IsIPAllocatedByMe(dnet, ep, ip) {
            return handler
        }
    }
    return nil
}

// DeleteDanmEp selects an ReleaseIPService implementation for the allocated IPv4 and IPv6 IP addresses
// in a given DanmEp and invokes them separately trying ot free up allocated IPs, after successfully freeing up
// the IPs, it deletes the DanmEp or returns error if unable to free any of the IPs or unable todelete DanmEp
func DeleteDanmEp(danmClient danmclientset.Interface, ep *danmtypes.DanmEp, dnet *danmtypes.DanmNet) error {
    if ! ipam.WasIpAllocatedByDanm(ep.Spec.Iface.Address, dnet.Spec.Options.Cidr) {
        if service := SelectAppropriateServiceImplementation(dnet, ep, ep.Spec.Iface.Address); service != nil {
            if err := service.ReleaseIP(dnet, ep, ep.Spec.Iface.Address); err != nil {
                return fmt.Errorf("unable to release ipv4 IP because: %s", err)
            }
        }
    }
    if ! ipam.WasIpAllocatedByDanm(ep.Spec.Iface.AddressIPv6, dnet.Spec.Options.Pool6.Cidr) {
        if service := SelectAppropriateServiceImplementation(dnet, ep, ep.Spec.Iface.AddressIPv6); service != nil {
            if err := service.ReleaseIP(dnet, ep, ep.Spec.Iface.AddressIPv6); err != nil {
                return fmt.Errorf("unable to release ipv6 IP because: %s", err)
            }
        }
    }
    return danmep.DeleteDanmEp(danmClient, ep, dnet)
}
