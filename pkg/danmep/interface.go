package danmep

import (
	"context"
	"fmt"
	"log"

	danmtypes "github.com/nokia/danm/crd/apis/danm/v1"
	danmclientset "github.com/nokia/danm/crd/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetSupportedReleaseIPHandlers gets all supported static CNI implementation that cleaner can use
// when deleting dangling DanmEps
func GetSupportedReleaseIPHandlers(danmClient danmclientset.Interface, dnet *danmtypes.DanmNet) []interface{} {
	return []interface{}{
		&danmReleaseIPServiceImpl{danmClient, dnet},
		&calicoReleaseIPServiceImpl{danmClient, dnet},
	}
}

// ReleaseIPInterface need to be implemented by every static CNI IPAM release plugin
// that want to invoked when Cleaner is cleaning up dangling DanmEps
type ReleaseIPInterface interface {
	IsIPAllocatedByMe(ip string, neType string) bool
	ReleaseIP(ip string) error
}

type releaseIPServiceImplBase struct {
	danmClient danmclientset.Interface
	dnet       *danmtypes.DanmNet
}

// DeleteDanmEp selects a ReleaseIPService implementation for allocated IPv4 and IPv6 IP addresses in a given DanmEp and invokes them to free up the allocated IPs in a remote control plane
// After successfully freeing up all IPs, it deletes the DanmEp as well
// Returns error if a registered service for the network was unable to free any of the IPs, or if the DanmEp could not be deleted
func DeleteDanmEp(danmClient danmclientset.Interface, ep *danmtypes.DanmEp, dnet *danmtypes.DanmNet) error {
	v4Error := releaseIP(danmClient, dnet, ep.Spec.NetworkType, ep.Spec.Iface.Address)
	v6Error := releaseIP(danmClient, dnet, ep.Spec.NetworkType, ep.Spec.Iface.AddressIPv6)
	if v4Error != nil || v6Error != nil {
		return fmt.Errorf("one of the registered IP releasing service failed with an error: V4: %s, V6: %s", v4Error, v6Error)
	}
	return danmClient.DanmV1().DanmEps(ep.ObjectMeta.Namespace).Delete(context.TODO(), ep.ObjectMeta.Name, metav1.DeleteOptions{})
}

func releaseIP(danmClient danmclientset.Interface, dnet *danmtypes.DanmNet, neType string, ip string) error {
	if ip == "" {
		return nil
	}
	service := selectReleaseIpServiceImplementation(danmClient, dnet, neType, ip)
	if service == nil {
		log.Println("WARNING: no releaseIP Service is found for IP:" + ip + " in network:" + dnet.ObjectMeta.Name + ", deleting Endpoint without additional clean-up!")
		return nil
	}
	err := service.ReleaseIP(ip)
	if err != nil {
		return err
	}
	return nil
}

// selectReleaseIpServiceImplementation looks up and returns the
// first suitable ReleaseIP service implementation for given IP address
// returns nil when no suitable implementation found
func selectReleaseIpServiceImplementation(danmClient danmclientset.Interface, dnet *danmtypes.DanmNet, neType string, ip string) ReleaseIPInterface {
	for _, impl := range GetSupportedReleaseIPHandlers(danmClient, dnet) {
		if service, ok := impl.(ReleaseIPInterface); ok && service.IsIPAllocatedByMe(ip, neType) {
			return service
		}
	}
	return nil
}
